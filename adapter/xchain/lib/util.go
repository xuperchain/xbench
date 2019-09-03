package lib

import (
//	"fmt"
	"os"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io/ioutil"
	"path/filepath"
	"encoding/base64"
	"strconv"
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

func CreateAcct() (*Acct, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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
	txs := SignTx(tx, from, from.Address, bcname)
	return txs
}

func Transfer(from *Acct, to string, bcname string, amount string) (*pb.CommonReply, error) {
	tx := FormatTx(from.Address)
	FormatTxOutput(tx, to, amount, "0")
	FormatTxInput(tx, bcname, from, from.Address)
	txs := SignTx(tx, from, from.Address, bcname)
	return PostTx(txs)
}

func TransferSplit(from *Acct, to string, bcname string, amount int) (*pb.CommonReply, error) {
	tx := FormatTx(from.Address)
	for i:=0; i<amount; i++ {
		FormatTxOutput(tx, to, "1", "0")
	}
	FormatTxInput(tx, bcname, from, from.Address)
	txs := SignTx(tx, from, from.Address, bcname)
	return PostTx(txs)
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
