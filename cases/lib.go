package cases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/xuperchain/crypto/core/hash"
	"github.com/xuperchain/xuper-sdk-go/account"
	"github.com/xuperchain/xupercore/lib/crypto/client"
	"github.com/xuperchain/xupercore/lib/utils"
	"github.com/xuperchain/xuperos/common/xupospb/pb"
	"github.com/xuperchain/xuperos/service/adapter/common"
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// 产生随机字符串
func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return b
}

func GenerateTx(response *pb.PreExecWithSelectUTXOResponse, ak *account.Account) *pb.Transaction {
	tx := &pb.Transaction{
		Version:   1,
		Coinbase:  false,
		Desc:      []byte(""),
		Nonce:     utils.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: ak.Address,

		TxInputs: response.GetResponse().UtxoInputs,
		TxOutputs: response.GetResponse().UtxoOutputs,

		ContractRequests: response.GetResponse().GetRequests(),
		TxInputsExt: response.GetResponse().GetInputs(),
		TxOutputsExt: response.GetResponse().GetOutputs(),
	}

	total := big.NewInt(0)
	for i := range tx.TxOutputs {
		amount := big.NewInt(0).SetBytes(tx.TxOutputs[i].GetAmount())
		total.Add(amount, total)
	}

	if response.GetResponse().GetGasUsed() > 0 {
		amount := big.NewInt(response.GetResponse().GetGasUsed())
		total.Add(amount, total)
		txFee := &pb.TxOutput{
			ToAddr: []byte("$"),
			Amount: amount.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txFee)
	}
	if total.Sign() > 0 {
		for _, utxo := range response.GetUtxoOutput().UtxoList {
			txInput := &pb.TxInput{
				RefTxid:   utxo.RefTxid,
				RefOffset: utxo.RefOffset,
				FromAddr:  utxo.ToAddr,
				Amount:    utxo.Amount,
			}
			tx.TxInputs = append(tx.TxInputs, txInput)
		}
		utxoTotal, _ := big.NewInt(0).SetString(response.GetUtxoOutput().TotalSelected, 10)
		if utxoTotal.Cmp(total) > 0 {
			delta := utxoTotal.Sub(utxoTotal, total)
			txCharge := &pb.TxOutput{
				ToAddr: []byte(ak.Address),
				Amount: delta.Bytes(),
			}
			tx.TxOutputs = append(tx.TxOutputs, txCharge)
		}
	}

	return tx
}

func PreExecWithSelectUTXO(ak *account.Account, amount int64) (*pb.PreExecWithSelectUTXOResponse, error) {
	content := hash.DoubleSha256([]byte("xuper" + ak.Address + strconv.FormatInt(amount, 10) + "true"))
	cryptoClient, err := client.CreateCryptoClient("default")
	if err != nil {
		return nil, err
	}
	privateKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJsonStr(ak.PrivateKey)
	sign, _ := cryptoClient.SignECDSA(privateKey, content)
	signInfo := &pb.SignatureInfo{
		PublicKey: ak.PublicKey,
		Sign: sign,
	}
	header := &pb.Header{
		Logid: utils.GenLogId(),
	}
	req := &pb.InvokeRPCRequest{
		Header: header,
		Bcname: "xuper",
		Requests: []*pb.InvokeRequest{},
		Initiator: ak.Address,
		AuthRequire: []string{ak.Address},
	}
	in := &pb.PreExecWithSelectUTXORequest{
		Header: header,
		Bcname: "xuper",
		Address: ak.Address,
		TotalAmount: amount,
		SignInfo: signInfo,
		NeedLock: true,
		Request: req,
	}

	return xchain.PreExecWithSelectUTXO(context.Background(), in)
}

//{
//  "module_name": "wasm",
//  "contract_name": "counter.wasm",
//  "method_name": "increase",
//  "args" : {
//    "key": "test",
//  }
//}
func PreExecWithSelectUTXOContract(request *pb.InvokeRequest, ak *account.Account) (*pb.PreExecWithSelectUTXOResponse, error) {
	authRequire := make([]string, 0, 1)
	accountName, ok := request.Args["account_name"]
	if ok {
		authRequire = append(authRequire, fmt.Sprintf("%s/%s", accountName, ak.Address))
	} else {
		authRequire = append(authRequire, ak.Address)
	}

	header := &pb.Header{
		Logid: utils.GenLogId(),
	}
	req := &pb.InvokeRPCRequest{
		Header: header,
		Bcname: "xuper",
		Requests: []*pb.InvokeRequest{request},
		Initiator: ak.Address,
		AuthRequire: authRequire,
	}

	address := ak.Address
	if request.ModuleName == "xkernel" && request.MethodName == "Deploy" {
		address = string(accountName)
	}
	content := hash.DoubleSha256([]byte("xuper" + address + "0" + "true"))
	cryptoClient, err := client.CreateCryptoClient("default")
	if err != nil {
		return nil, err
	}
	privateKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJsonStr(ak.PrivateKey)
	sign, _ := cryptoClient.SignECDSA(privateKey, content)
	signInfo := &pb.SignatureInfo{
		PublicKey: ak.PublicKey,
		Sign: sign,
	}
	in := &pb.PreExecWithSelectUTXORequest{
		Header: header,
		Bcname: "xuper",
		Address: address,
		TotalAmount: 0,
		SignInfo: signInfo,
		NeedLock: true, // TODO: true
		Request: req,
	}

	return xchain.PreExecWithSelectUTXO(context.Background(), in)
}

func SignTx(tx *pb.Transaction, from *account.Account, account string) *pb.Transaction {
	if account != "" {
		tx.AuthRequire = append(tx.AuthRequire, account + "/" + from.Address)
	} else {
		tx.AuthRequire = append(tx.AuthRequire, from.Address)
	}

	cryptoClient, _ := client.CreateCryptoClient("default")
	signTx, _ := common.ComputeTxSign(cryptoClient, tx, []byte(from.PrivateKey))
	signInfo := &pb.SignatureInfo{
		PublicKey: from.PublicKey,
		Sign: signTx,
	}
	tx.InitiatorSigns = append(tx.InitiatorSigns, signInfo)
	tx.AuthRequireSigns = append(tx.AuthRequireSigns, signInfo)
	tx.Txid, _ = common.MakeTxId(tx)
	return tx
}


func SplitOutputs(account string, total *big.Int, num int64) ([]*pb.TxOutput, error) {
	if big.NewInt(num).Cmp(total) == 1 {
		return nil, errors.New("split utxo <= balance required")
	}

	amount := big.NewInt(0)
	amount.Div(total, big.NewInt(num))

	output := pb.TxOutput{}
	output.Amount = amount.Bytes()
	output.ToAddr = []byte(account)

	rest := total
	txOutputs := make([]*pb.TxOutput, 0, num+1)
	for i := int64(1); i < num && rest.Cmp(amount) == 1; i++ {
		tmpOutput := output
		txOutputs = append(txOutputs, &tmpOutput)
		rest.Sub(rest, amount)
	}
	output.Amount = rest.Bytes()
	txOutputs = append(txOutputs, &output)
	return txOutputs, nil
}


func PrintTx(tx *pb.Transaction) error {
	t := FromPBTx(tx)
	output, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}