package cases

import (
	"errors"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Query struct {
	common.TestCase
}

func (q Query) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.Connect(env.Host)
	//Accts = lib.CreateTestClients(parallel, env.Host)
	for i:=0; i<parallel; i++ {
		Accts[i], _ = lib.CreateAcct()
	}
	Bank = lib.InitBankAcct("")
	log.INFO.Printf("prepare tokens of test accounts ...")
	for i := range Accts {
		rsp, err := lib.Transfer(Bank, Accts[i].Address, env.Chain, "10")
		if rsp.Header.Error != 0 || err != nil {
			log.ERROR.Printf("prepare tokens error: %#v, rsp: %#v", err, rsp)
			return errors.New("init token error")
		}
	}
	return nil
}

func (q Query) Run(seq int, args ...interface{}) error {
	acct := Accts[seq]
	rsp, err := lib.GetBalance(acct.Address, "xuper")
	if rsp.Header.Error != 0 || err != nil {
		return errors.New("run query error")
	}
	return err
}

func (q Query) End(args ...interface{}) error {
	return nil
}
