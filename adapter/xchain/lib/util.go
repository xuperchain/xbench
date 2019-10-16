package lib

import (
	"fmt"
	"os"
	"time"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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

func GenProfTx(from *Acct, to string, bcname string) *pb.TxStatus {
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, to, "1", "0")
	FormatTxInput(tx, bcname, from, from.Address)
	FormatTxReserved(tx, from.Address, bcname)
	txs := SignTx(tx, from, "", bcname)
	return txs
}

func WaitTx(retry int, txid string, bcname string) bool {
	for i := 0; i < retry; i++ {
		status := QueryTx(txid, bcname)
		if status.Status == 2 {
			return true
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
	return false
}

func Transfer(from *Acct, to string, bcname string, amount string) (*pb.CommonReply, string, error) {
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, to, amount, "0")
	FormatTxInput(tx, bcname, from, from.Address)
	FormatTxReserved(tx, from.Address, bcname)
	txs := SignTx(tx, from, "", bcname)
	txid := fmt.Sprintf("%x", txs.Txid)
	rsp, err := PostTx(txs)
	return rsp, txid, err
}

func Transfer2 (from *Acct, to string, bcname string, amount string) (*pb.CommonReply, string, error) {
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, to, amount, "0")
	FormatTxUtxoPreExec(tx, bcname, from)
	txs := SignTx(tx, from, "", bcname)
	txid := fmt.Sprintf("%x", txs.Txid)
	rsp, err := PostTx(txs)
	return rsp, txid, err
}

func TransferSplit(from *Acct, to string, bcname string, amount int) (*pb.CommonReply, string, error) {
	tx := FormatTx(from.Address)
	for i:=0; i<amount; i++ {
		FormatTxOutput(tx, to, "1", "0")
	}
	FormatTxInput(tx, bcname, from, from.Address)
	FormatTxReserved(tx, from.Address, bcname)
	txs := SignTx(tx, from, "", bcname)
	txid := fmt.Sprintf("%x", txs.Txid)
	rsp, err := PostTx(txs)
	return rsp, txid, err
}

func NewContractAcct(from *Acct, name string, bcname string) (*pb.CommonReply, error) {
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
	rsp, req, _ := PreExec(args, "xkernel", "NewAccount", bcname, "", from.Address)
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, "$", strconv.FormatInt(rsp.GasUsed, 10), "0")
	FormatTxInput(tx, bcname, from, from.Address)
	FormatTxExt(tx, rsp, req)
	txs := SignTx(tx, from, "", bcname)
	return PostTx(txs)
}

func DeployContract(from *Acct, code string, name string, contract string, bcname string) (*pb.CommonReply, error) {
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
	rsp, req, err := PreExec(args, "xkernel", "Deploy", bcname, "", from.Address)
	if err != nil {
		return nil, err
	}
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, "$", strconv.FormatInt(rsp.GasUsed, 10), "0")
	FormatTxInput(tx, bcname, from, name)
	FormatTxExt(tx, rsp, req)
	txs := SignTx(tx, from, name, bcname)
	return PostTx(txs)
}

func InvokeContract(from *Acct, contract string, bcname string, method string, key string) (*pb.CommonReply, error) {
	args := make(map[string][]byte)
	args["key"] = []byte(key)
	rsp, req, err := PreExec(args, "wasm", method, bcname, contract, from.Address)
	if err != nil {
		return nil, err
	}
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, "$", strconv.FormatInt(rsp.GasUsed, 10), "0")
	FormatTxInput(tx, bcname, from, from.Address)
	FormatTxExt(tx, rsp, req)
	txs := SignTx(tx, from, "", bcname)
	return PostTx(txs)
}

func QueryContract(from *Acct, contract string, bcname string, method string, key string) (*pb.InvokeResponse, []*pb.InvokeRequest, error) {
	args := make(map[string][]byte)
	args["key"] = []byte(key)
	rsp, reqs, err := PreExec(args, "wasm", method, bcname, contract, "")
	return rsp, reqs, err
}

func InitIdentity(from *Acct, bcname string, accts []string) (*pb.CommonReply, error) {
	args := make(map[string][]byte)
	args["aks"] = []byte(strings.Join(accts, ","))
	rsp, req, err := PreExec(args, "wasm", "register_aks", bcname, "identity", from.Address)
	if err != nil {
		return nil, err
	}
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, "$", strconv.FormatInt(rsp.GasUsed, 10), "0")
	FormatTxInput(tx, bcname, from, from.Address)
	FormatTxExt(tx, rsp, req)
	txs := SignTx(tx, from, "", bcname)
	return PostTx(txs)
}
