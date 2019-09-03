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
)

func (i Invoke) Init(args ...interface{}) error {
	env := args[1].(common.TestEnv)
	lib.Connect(env.Host)
	Bank = lib.InitBankAcct("")
	log.INFO.Printf("check contract account ...")
	account := fmt.Sprintf("XC%s@xuper", acct)
	rsp := lib.QueryACL(env.Chain, account)
	if !rsp.Confirmed {
		lib.NewContractAcct(Bank, acct, env.Chain)
	}
	lib.Transfer(Bank, account, env.Chain, "10000000")
	log.INFO.Printf("check counter contract ...")
	_, _, err := lib.QueryContract(Bank, contract, env.Chain, "get", "key_1")
	if err != nil {
		lib.DeployContract(Bank, contractpath, account, contract, env.Chain)
	}
	log.INFO.Printf("prepare done %s on %s", account, contract)
	return nil
}

func (i Invoke) Run(seq int, args ...interface{}) error {
	k := fmt.Sprintf("key_%d", seq)
	rsp, err := lib.InvokeContract(Bank, contract, "xuper", "increase", k)
	if err != nil || rsp.Header.Error != 0 {
		log.ERROR.Printf("err on invoke %#v", rsp.Header)
		return errors.New("invoke contract error")
	}
	return nil
}

func (i Invoke) End(args ...interface{}) error {
	return nil
}
