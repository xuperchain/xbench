package cases

import (
	"errors"
	"fmt"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Invoke struct {
	common.TestCase
}

var (
	acct         = "1123581321345589"
	contract     = "counter"
	contractpath = "data/counter"
)

// Init implements the comm.IcaseFace
func (i Invoke) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
	txid := ""
	for i := 0; i <= parallel-1 && len(Clis) < parallel; i++ {
		cli := lib.Conn(env.Host, env.Chain)
		Clis = append(Clis, cli)
	}
	Bank = lib.InitBankAcct("")
	log.INFO.Printf("check contract account ...")
	account := fmt.Sprintf("XC%s@%s", acct, env.Chain)
	status, err := Clis[0].QueryACL(account)
	if err != nil || status == nil {
		log.ERROR.Printf("queryacl error")
		return err
	}
	if !status.Confirmed {
		log.INFO.Printf("new contract account ...")
		_, txid, _ = lib.NewContractAcct(Bank, acct, Clis[0])
		lib.WaitConfirm(txid, 5, Clis[0])
	}
	log.INFO.Printf("check counter contract ...")
	_, _, err = lib.QueryContract(Bank, contract, "get", "key_0", Clis[0])
	if err != nil {
		_, txid, _ = lib.Trans(Bank, account, "10000000", Clis[0])
		lib.WaitConfirm(txid, 5, Clis[0])
		_, txid, _ = lib.DeployContract(Bank, contractpath, account, contract, Clis[0])
		lib.WaitConfirm(txid, 5, Clis[0])
	}
	log.INFO.Printf("prepare done %s on %s", account, contract)
	return nil
}

func (i Invoke) Run(seq int, args ...interface{}) error {
	k := fmt.Sprintf("key_%d", seq)
	rsp, _, err := lib.InvokeContract(Bank, contract, "Increase", k, Clis[seq])
	if err != nil || rsp.Header.Error != 0 {
		log.ERROR.Printf("err on invoke %#v", rsp.Header)
		return errors.New("invoke contract error")
	}
	return nil
}

func (i Invoke) End(args ...interface{}) error {
	return nil
}
