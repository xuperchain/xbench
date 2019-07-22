package xchain

import (
	"context"
	"encoding/hex"
	"math/big"
	"io/ioutil"
	"strconv"
	"time"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
	"github.com/xuperchain/xuperunion/crypto/client"
	"google.golang.org/grpc"
	"github.com/golang/protobuf/proto"
)

var (
	conn *grpc.ClientConn
	cli pb.XchainClient
)

func Connect(host string) {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithMaxMsgSize(64<<20-1))
	conn, _ = grpc.Dial(host, opts...)
	cli = pb.NewXchainClient(conn)
}

func header() *pb.Header {
	out := &pb.Header{
		Logid: global.Glogid(),
	}
	return out
}

func GetBalance(addr string, bcname string) (*pb.AddressStatus, error) {
	bc := &pb.TokenDetail{
		Bcname: bcname,
	}
	in := &pb.AddressStatus{
		Header: header(),
		Address: addr,
		Bcs: []*pb.TokenDetail{bc},
	}
	out, err := cli.GetBalance(context.Background(), in)
	return out, err
}

func GetFrozenBalance(addr string, bcname string) *pb.AddressStatus {
	bc := &pb.TokenDetail{
		Bcname: bcname,
	}
	in := &pb.AddressStatus{
		Header: header(),
		Address: addr,
		Bcs: []*pb.TokenDetail{bc},
	}
	out, _ := cli.GetFrozenBalance(context.Background(), in)
	return out
}

func GetBlock(blockid string, bcname string) *pb.Block {
	id, _ := hex.DecodeString(blockid)
	in := &pb.BlockID{
		Header: header(),
		Bcname: bcname,
		Blockid: id,
		NeedContent: true,
	}
	out, _ := cli.GetBlock(context.Background(), in)
	return out
}

func QueryTx(txid string, bcname string) *pb.TxStatus {
	tx , _ := hex.DecodeString(txid)
	in := &pb.TxStatus{
		Header: header(),
		Bcname: bcname,
		Txid: tx,
	}
	out, _ := cli.QueryTx(context.Background(), in)
	return out
}

func QueryACL(bcname string, acct string) *pb.AclStatus {
	in := &pb.AclStatus{
		Header: header(),
		Bcname: bcname,
		AccountName: acct,
	}
	out, _ := cli.QueryACL(context.Background(), in)
	return out
}

func GetBlockChains() *pb.BlockChains {
	in := &pb.CommonIn{
		Header: header(),
	}
	out, _ := cli.GetBlockChains(context.Background(), in)
	return out
}

func GetSystemStatus() *pb.SystemsStatusReply {
	in := &pb.CommonIn{
		Header: header(),
	}
	out, _ := cli.GetSystemStatus(context.Background(), in)
	return out
}

func SelectUTXO(f *Acct, bcname string, need string, name string) (*pb.UtxoOutput, error) {
	in := &pb.UtxoInput{
		Header: header(),
		Bcname: bcname,
		Address: name,
		Publickey: f.Pub,
		TotalNeed: need,
		NeedLock: true,
	}
	out, err := cli.SelectUTXO(context.Background(), in)
	if err != nil {
		log.ERROR.Printf("select utxo error %#v", err)
		return nil, err
	}
	return out, nil
}

func FormatTx(from string) *pb.Transaction {
	tx := &pb.Transaction{
		Version: 1,
		Coinbase: false,
		Desc: []byte(""),
		Nonce: global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: from,
	}
	return tx
}

func FormatTxOutput(tx *pb.Transaction, to string, amount string, frozen string) {
	amt, _ := big.NewInt(0).SetString(amount, 10)
	txout := &pb.TxOutput{
		ToAddr: []byte(to),
		Amount: amt.Bytes(),
	}
	if frozen != "0" {
		frz, _ := strconv.ParseInt(frozen, 10, 64)
		txout.FrozenHeight = frz
	}
	tx.TxOutputs = append(tx.TxOutputs, txout)
}

func FormatTxInput(tx *pb.Transaction, bcname string, from *Acct, name string) {
	total := big.NewInt(0)
	for i := range(tx.TxOutputs) {
		amt := big.NewInt(0).SetBytes(tx.TxOutputs[i].GetAmount())
		total.Add(amt, total)
	}
	utxoRes, _ := SelectUTXO(from, bcname, total.String(), name)
	for _, utxo := range utxoRes.UtxoList {
		txInput := &pb.TxInput{
			RefTxid: utxo.RefTxid,
			RefOffset: utxo.RefOffset,
			FromAddr: utxo.ToAddr,
			Amount: utxo.Amount,
		}
		tx.TxInputs = append(tx.TxInputs, txInput)
	}
	utxoTotal, _ := big.NewInt(0).SetString(utxoRes.TotalSelected, 10)
	// fill the charge
	if utxoTotal.Cmp(total) > 0 {
		delta := utxoTotal.Sub(utxoTotal, total)
		txCharge := &pb.TxOutput{
			ToAddr: []byte(name),
			Amount: delta.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txCharge)
	}
}

func FormatTxExt(tx *pb.Transaction, rsp *pb.InvokeResponse, req *pb.InvokeRequest) {
	tx.TxInputsExt = rsp.GetInputs()
	tx.TxOutputsExt = rsp.GetOutputs()
	tx.ContractRequest = req
}

func SignTx(tx *pb.Transaction, from *Acct, name string, bcname string) *pb.TxStatus{
	if name != "" {
		tx.AuthRequire = append(tx.AuthRequire, name + "/" + from.Address)
	} else {
		tx.AuthRequire = append(tx.AuthRequire, from.Address)
	}
	cryptoClient, _ := client.CreateCryptoClient("default")
	signTx, _ := txhash.ProcessSignTx(cryptoClient, tx, []byte(from.Pri))
	signInfo := &pb.SignatureInfo{
		PublicKey: from.Pub,
		Sign: signTx,
	}
	tx.InitiatorSigns = append(tx.InitiatorSigns, signInfo)
	tx.AuthRequireSigns = append(tx.AuthRequireSigns, signInfo)
	tx.Txid, _ = txhash.MakeTransactionID(tx)
	txStatus := &pb.TxStatus{
		Bcname: bcname,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx: tx,
	}
	txStatus.Txid = tx.Txid
	return txStatus
}

func GenTx(f *Acct, taddr string, bcname string, amount string, fee string,
	desc string, frozen string, issplit bool, ivr *pb.InvokeResponse,
	ivq *pb.InvokeRequest) (*pb.TxStatus, error) {
	// build a Tx struct
	tx := &pb.Transaction{
		Version: 1,
		Coinbase: false,
		Desc: []byte(desc),
		Nonce: global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: f.Address,
	}

	// fill the target output
	froz, _ := strconv.ParseInt(frozen, 10, 64)
	amt, _ := big.NewInt(0).SetString(amount, 10)
	fe, _ := big.NewInt(0).SetString(fee, 10)
	total := big.NewInt(0)
	total.Add(amt, fe)
	if !issplit {
		txOutput := &pb.TxOutput{
			ToAddr: []byte(taddr),
			Amount: amt.Bytes(),
			FrozenHeight: froz,
		}
		tx.TxOutputs = append(tx.TxOutputs, txOutput)
		if fee != "" {
			feeOutput := &pb.TxOutput{
				ToAddr: []byte("$"),
				Amount: fe.Bytes(),
			}
			tx.TxOutputs = append(tx.TxOutputs, feeOutput)
		}
	} else {
		nib, _ := big.NewInt(0).SetString("1", 10)
		cnt, _ := strconv.Atoi(amount)
		for i:=0; i<cnt; i++ {
			txOutput := &pb.TxOutput{
				ToAddr: []byte(taddr),
				Amount: nib.Bytes(),
				FrozenHeight: froz,
			}
			tx.TxOutputs = append(tx.TxOutputs, txOutput)
		}
	}
	// select the token of fromer
	utxoRes, err := SelectUTXO(f, bcname, total.String(), f.Address)
	if err != nil {
		return nil, err
	}

	for _, utxo := range utxoRes.UtxoList {
		txInput := &pb.TxInput{
			RefTxid: utxo.RefTxid,
			RefOffset: utxo.RefOffset,
			FromAddr: utxo.ToAddr,
			Amount: utxo.Amount,
		}
		tx.TxInputs = append(tx.TxInputs, txInput)
	}
	utxoTotal, _ := big.NewInt(0).SetString(utxoRes.TotalSelected, 10)

	// fill the charge
	if utxoTotal.Cmp(total) > 0 {
		delta := utxoTotal.Sub(utxoTotal, total)
		txCharge := &pb.TxOutput{
			ToAddr: []byte(f.Address),
			Amount: delta.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txCharge)
	}

	if ivr != nil && ivq != nil {
		tx.TxInputsExt = ivr.GetInputs()
		tx.TxOutputsExt = ivr.GetOutputs()
		tx.ContractRequest = ivq
		tx.AuthRequire = []string{"XC1234612345123451@xuper/" + f.Address}
	} else {
		tx.AuthRequire = []string{f.Address + "/" + f.Address}
	}

	txStatus := &pb.TxStatus{
		Bcname: bcname,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx: tx,
	}

	// sign the Tx
	cryptoClient, err := client.CreateCryptoClient("default")
	if err != nil {
		return nil, err
	}
	signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(f.Pri))
	if err != nil {
		return nil, err
	}
	signInfo := &pb.SignatureInfo{
		PublicKey: f.Pub,
		Sign: signTx,
	}
	txStatus.Tx.InitiatorSigns = append(txStatus.Tx.InitiatorSigns, signInfo)
	txStatus.Tx.AuthRequireSigns = append(txStatus.Tx.AuthRequireSigns, signInfo)
	txStatus.Tx.Txid, _ = txhash.MakeTransactionID(txStatus.Tx)
	txStatus.Txid = txStatus.Tx.Txid

	return txStatus, nil
}

func PostTx(txstatus *pb.TxStatus) (*pb.CommonReply, error) {
	out, err := cli.PostTx(context.Background(), txstatus)
	return out, err
}

func PreExec(args map[string][]byte, module string, method string, bcname string, contract string) (*pb.InvokeResponse, *pb.InvokeRequest, error) {
	req := &pb.InvokeRequest{
		ModuleName: module,
		MethodName: method,
		Args: args,
	}
	if contract != "" && module != "xkernel" {
		req.ContractName = contract
	}
	in := &pb.InvokeRPCRequest{
		Bcname: bcname,
		Request: req,
		Header: header(),
	}
	out, err := cli.PreExec(context.Background(), in)
	return out.GetResponse(), req, err
}

// to new contract account
func GenPreExeRes(name string, addr string, chain string) (*pb.InvokeRPCResponse, *pb.InvokeRequest) {
	args := make(map[string][]byte)
	args["account_name"] = []byte(name)
	acl := `{
		"pm": {
            "rule": 1,
            "acceptValue": 1.0
        },
        "aksWeight": {
            "` + addr + `": 1.0
        }
	}`
	args["acl"] = []byte(acl)
	ivq := &pb.InvokeRequest{
		ModuleName: "xkernel",
		MethodName: "NewAccount",
		Args: args,
	}
//	in := []*pb.InvokeRequest{ivq}
	rpcin := &pb.InvokeRPCRequest{
		Bcname: chain,
		Request: ivq,
		Header: header(),
	}
	rpcout, _ := cli.PreExec(context.Background(), rpcin)
	return rpcout, ivq
}

func GenContractExeRes(initargs string, acctname string, contract string, code string,
	lang string, chain string) (*pb.InvokeRPCResponse, *pb.InvokeRequest) {
	// contra
	args := make(map[string][]byte)
	args["account_name"] = []byte(acctname)
	args["contract_name"] = []byte(contract)
	desc := &pb.WasmCodeDesc{
		Runtime: lang,
	}
	buf , _ := proto.Marshal(desc)
	args["contract_desc"] = buf
	codebuf, _ := ioutil.ReadFile(code)
	args["contract_code"] = codebuf
	args["init_args"] = []byte(initargs)
	ivq := &pb.InvokeRequest{
		ModuleName: "xkernel",
		MethodName: "Deploy",
		Args: args,
	}
//	in := []*pb.InvokeRequest{ivq}
	rpcin := &pb.InvokeRPCRequest{
		Bcname: chain,
		Request: ivq,
		Header: header(),
	}
	rpcout, _ := cli.PreExec(context.Background(), rpcin)
	return rpcout, ivq
}
