package xchain

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"path/filepath"
	"encoding/base64"
	"strconv"
	"strings"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperunion/pb"
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

func InitBankAcct() *Acct {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
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

func CreateTestClients(num int, host string) map[int]*Acct {
	accts := map[int]*Acct{}
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	testKeysPath := filepath.Join(dir, "../data/testkeys/")
	if _, err := os.Stat(testKeysPath); os.IsNotExist(err) {
        os.Mkdir(testKeysPath, 0755)
    }
	for i:=0; i<num; i+=1 {
		tpath := filepath.Join(testKeysPath, strconv.Itoa(i))
		args := fmt.Sprintf("account newkeys -o %s", tpath)
		if _, e := os.Stat(tpath); os.IsNotExist(e) {
			RunCliCmd(args, host)
		}
		acct := &Acct{
			Address: getFileContent(tpath + "/address"),
			Pub: getFileContent(tpath + "/public.key"),
			Pri: getFileContent(tpath + "/private.key"),
		}
		accts[i] = acct
	}
	return accts
}

func RunCliCmd(args string, host string) string {
	var out bytes.Buffer
	f := strings.Fields(args + " -H " + host)
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	cmd := exec.Command(dir + "/xchain-cli", f...)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
	    return ""
	}
	return out.String()
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
	rsp, req, _ := PreExec(args, "xkernel", "NewAccount", bcname, "")
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
	rsp, req, err := PreExec(args, "xkernel", "Deploy", bcname, "")
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
	rsp, req, err := PreExec(args, "wasm", method, bcname, contract)
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

func QueryContract(from *Acct, contract string, bcname string, method string, key string) (*pb.InvokeResponse, *pb.InvokeRequest, error) {
	args := make(map[string][]byte)
	args["key"] = []byte(key)
	rsp, req, err := PreExec(args, "wasm", method, bcname, contract)
	return rsp, req, err
}
