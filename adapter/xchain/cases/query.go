package cases

import (
	"errors"
	"fmt"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Query struct {
	common.TestCase
}

var (
	qacct = "1123581321345589"
	qcontract = "proftestc"
	qcontractpath = "data/counter"
	qchainname = ""
)

func (q Query) Init(args ...interface{}) error {
	env := args[1].(common.TestEnv)
	qchainname = env.Chain
	lib.Connect(env.Host, env.Nodes, env.Crypto)
	Bank = lib.InitBankAcct("")
	log.INFO.Printf("check contract account ...")
	account := fmt.Sprintf("XC%s@%s", qacct, qchainname)
	rsp := lib.QueryACL(qchainname, account)
	if !rsp.Confirmed {
		lib.NewContractAcct(Bank, qacct, qchainname)
	}
	lib.Transfer(Bank, account, env.Chain, "10000000")
	log.INFO.Printf("check counter contract ...")
	_, _, err := lib.QueryContract(Bank, contract, qchainname, "get", "key_0")
	if err != nil {
		lib.DeployContract(Bank, contractpath, account, qcontract, qchainname)
	}
	log.INFO.Printf("prepare done %s on %s", account, qcontract)
	lib.InvokeContract(Bank, qcontract, qchainname, "increase", "key_0")
	return nil
}

func (q Query) Run(seq int, args ...interface{}) error {
	rsp, _, err := lib.QueryContract(Bank, qcontract, qchainname, "get", "key_0")
	if rsp == nil || err != nil {
		return errors.New("run query error")
	}
	return err
}

func (q Query) End(args ...interface{}) error {
	return nil
}
