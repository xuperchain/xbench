package lib

import (
	"os"
	"time"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"io/ioutil"
	"path/filepath"
	"encoding/base64"
	"strconv"
	"strings"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/crypto/account"
	"github.com/golang/protobuf/proto"
)

type Acct struct {
	Address string
	Pub string
	Pri string
}

func getFileContent(file string) string {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.ERROR.Printf("read file error: %s", err)
		return ""
	}
	return string(f)
}

func InitBankAcct(dir string) *Acct {
	if dir == "" {
		dir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	keyPath := filepath.Join(dir, "../data/keys")
	addr := getFileContent(keyPath + "/address")
	pubkey := getFileContent(keyPath + "/public.key")
	scrkey := getFileContent(keyPath + "/private.key")
	acct := &Acct{
		Address: addr,
		Pub: pubkey,
		Pri: scrkey,
	}
	return acct
}

func CreateAcct(cryptotype string) (*Acct, error) {
	curve := elliptic.P256()
	if cryptotype == "schnorr" {
		curve.Params().Name = "P-256-SN"
	}
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	pri, _ := account.GetEcdsaPrivateKeyJSONFormat(privateKey)
	pub, _ := account.GetEcdsaPublicKeyJSONFormat(privateKey)
	addr, err := account.GetAddressFromPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	acct := &Acct{
		Address: addr,
		Pub: pub,
		Pri: pri,
	}
	return acct, nil
}

func ProfTx(from *Acct, to string, cli *Client) *pb.TxStatus {
	tx := FormatTx(from.Address)
	out, _ := cli.PreExecWithSelectUTXO(from, 1)
	FormatOutput(tx, to, "1", "0")
	FormatInputPreExec(tx, from, out)
	return cli.SignTx(tx, from, "")
}

func Transplit(from *Acct, to string, amount int, cli *Client) (*pb.CommonReply, string, error) {
	tx := FormatTx(from.Address)
	out, _ := cli.PreExecWithSelectUTXO(from, int64(amount))
	for i:=0; i<=amount-1; i++ {
		FormatOutput(tx, to, "1", "0")
	}
	FormatInputPreExec(tx, from, out)
	txs := cli.SignTx(tx, from, "")
	return cli.PostTx(txs)
}

func Trans(from *Acct, to string, amount string, cli *Client) (*pb.CommonReply, string, error) {
	tx := FormatTx(from.Address)
	need, _ := strconv.Atoi(amount)
	out, _ := cli.PreExecWithSelectUTXO(from, int64(need))
	FormatOutput(tx, to, amount, "0")
	FormatInputPreExec(tx, from, out)
	txs := cli.SignTx(tx, from, "")
	return cli.PostTx(txs)
}

func NewContractAcct(from *Acct, name string, cli *Client) (*pb.CommonReply, string, error) {
	args := make(map[string][]byte)
	args["account_name"] = []byte(name)
	acl := `{
		"pm": {
            "rule": 1,
            "acceptValue": 1.0
        },
        "aksWeight": {
            "` + from.Address + `": 1.0
        }
	}`
	args["acl"] = []byte(acl)
	out, _ := cli.PreExecWithSelectUTXOContract(from, args, "xkernel", "NewAccount", "")
	tx := FormatTx(from.Address)
	FormatOutput(tx, from.Address, "0", "0")
	FormatInputPreExec(tx, from, out)
	txs := cli.SignTx(tx, from, name)
	return cli.PostTx(txs)
}

func DeployContract(from *Acct, code string, name string, contract string, cli *Client) (*pb.CommonReply, string, error) {
	args := make(map[string][]byte)
	args["account_name"] = []byte(name)
	args["contract_name"] = []byte(contract)
	desc := &pb.WasmCodeDesc{
		Runtime: "c",
	}
	buf , _ := proto.Marshal(desc)
	args["contract_desc"] = buf
	source, _ := ioutil.ReadFile(code)
	args["contract_code"] = source
	iarg := `{"creator":"` + base64.StdEncoding.EncodeToString([]byte("xchain")) + `"}`
	args["init_args"] = []byte(iarg)
	out, _ := cli.PreExecWithSelectUTXOContract(from, args, "xkernel", "Deploy", "")
	tx := FormatTx(from.Address)
	FormatInputPreExec(tx, from, out)
	txs := cli.SignTx(tx, from, name)
	return cli.PostTx(txs)
}

func InvokeContract(from *Acct, contract string, method string, key string, cli *Client) (*pb.CommonReply, string, error) {
	args := make(map[string][]byte)
	args["key"] = []byte(key)
	out, _ := cli.PreExecWithSelectUTXOContract(from, args, "wasm", method, contract)
	tx := FormatTx(from.Address)
	FormatInputPreExec(tx, from, out)
	txs := cli.SignTx(tx, from, "")
	return cli.PostTx(txs)
}

func InitIdentity(from *Acct, accts []string, cli *Client) (*pb.CommonReply, string, error) {

	args := make(map[string][]byte)
	args["aks"] = []byte(strings.Join(accts, ","))
	out, _ := cli.PreExecWithSelectUTXOContract(from, args, "wasm", "register_aks", "unified_check")
	tx := FormatTx(from.Address)
	FormatInputPreExec(tx, from, out)
	txs := cli.SignTx(tx, from, "")
	return cli.PostTx(txs)
}

func QueryContract(from *Acct, contract string, method string, key string, cli *Client) (*pb.InvokeResponse, []*pb.InvokeRequest, error) {
	args := make(map[string][]byte)
	args["key"] = []byte(key)
	return cli.PreExec(args, "wasm", method, contract, from.Address)
}

func FormatTx(from string) *pb.Transaction {
	return &pb.Transaction{
		Version: 1,
		Coinbase: false,
		Desc: []byte(""),
		Nonce: nonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: from,
	}
}

func FormatOutput(tx *pb.Transaction, to string, amount string, frozen string) {
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

func FormatRelayInput(tx *pb.Transaction, relayid string, rsp *pb.InvokeResponse) {
	tx.ContractRequests = rsp.GetRequests()
	tx.TxInputsExt = rsp.GetInputs()
	tx.TxOutputsExt = rsp.GetOutputs()
	refid, _ := hex.DecodeString(relayid)
	txOutput := tx.TxOutputs[0]
	txInput := &pb.TxInput{
		RefTxid: refid,
		RefOffset: 0,
		FromAddr: txOutput.ToAddr,
		Amount: txOutput.Amount,
	}
	tx.TxInputs = append(tx.TxInputs, txInput)
}

func FormatInputPreExec(tx *pb.Transaction, from *Acct, rsp *pb.PreExecWithSelectUTXOResponse) {
	total := big.NewInt(0)
	for i := range(tx.TxOutputs) {
		amt := big.NewInt(0).SetBytes(tx.TxOutputs[i].GetAmount())
		total.Add(amt, total)
	}
	tx.ContractRequests = rsp.GetResponse().GetRequests()
	tx.TxInputsExt = rsp.GetResponse().GetInputs()
	tx.TxOutputsExt = rsp.GetResponse().GetOutputs()
	if rsp.GetResponse().GetGasUsed() > 0 {
		amt := big.NewInt(rsp.GetResponse().GetGasUsed())
		total.Add(amt, total)
		txFee := &pb.TxOutput{
			ToAddr: []byte("$"),
			Amount: amt.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txFee)
	}
	if total.Sign() > 0 {
		for _, utxo := range rsp.GetUtxoOutput().UtxoList {
			txInput := &pb.TxInput{
				RefTxid: utxo.RefTxid,
				RefOffset: utxo.RefOffset,
				FromAddr: utxo.ToAddr,
				Amount: utxo.Amount,
			}
			tx.TxInputs = append(tx.TxInputs, txInput)
		}
		utxoTotal, _ := big.NewInt(0).SetString(rsp.GetUtxoOutput().TotalSelected, 10)
		if utxoTotal.Cmp(total) > 0 {
			delta := utxoTotal.Sub(utxoTotal, total)
			txCharge := &pb.TxOutput{
				ToAddr: []byte(from.Address),
				Amount: delta.Bytes(),
			}
			tx.TxOutputs = append(tx.TxOutputs, txCharge)
		}
	}
}
