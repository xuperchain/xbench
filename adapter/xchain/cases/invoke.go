package cases

import (
	"fmt"
	"errors"
//	"strings"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Invoke struct {
	common.TestCase
}

var (
	acct = "1123581321345589"
	contract = "proftestc"
	contractpath = "data/counter"
	chainname = ""
)

func (i Invoke) Init(args ...interface{}) error {
	env := args[1].(common.TestEnv)
	chainname = env.Chain
	lib.Connect(env.Host)
	Bank = lib.InitBankAcct("")
	log.INFO.Printf("check contract account ...")
	account := fmt.Sprintf("XC%s@%s", acct, chainname)
	rsp := lib.QueryACL(chainname, account)
	if !rsp.Confirmed {
		lib.NewContractAcct(Bank, acct, chainname)
	}
	lib.Transfer(Bank, account, env.Chain, "10000000")
	log.INFO.Printf("check counter contract ...")
	_, _, err := lib.QueryContract(Bank, contract, chainname, "get", "key_1")
	if err != nil {
		lib.DeployContract(Bank, contractpath, account, contract, chainname)
	}
	log.INFO.Printf("prepare done %s on %s", account, contract)
	return nil
}

func (i Invoke) Run(seq int, args ...interface{}) error {
	k := fmt.Sprintf("key_%d", seq)
	rsp, err := lib.InvokeContract(Bank, contract, chainname, "increase", k)
	if err != nil || rsp.Header.Error != 0 {
		log.ERROR.Printf("err on invoke %#v", rsp.Header)
		return errors.New("invoke contract error")
	}
	return nil
}

func (i Invoke) End(args ...interface{}) error {
	return nil
}
