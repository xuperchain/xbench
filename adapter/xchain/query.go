package xchain

import (
	"time"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Query struct {
	common.TestCase
}

var (
	bank = &Acct{}
	accts = map[int]*Acct{}
)

func (q Query) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	Connect(env.Host)
	accts = CreateTestClients(parallel, env.Host)
	bank = InitBankAcct()
	log.INFO.Printf("prepare tokens of test accounts ...")
	for i := range accts {
		txs, err := GenTx(bank, accts[i].Address, env.Chain, "10", "", "0")
		if err != nil {
			log.ERROR.Printf("prepare tokens error: %#v", err)
			return err
		}
		PostTx(txs)
	}
	time.Sleep(4 * time.Second)
	return nil
}

func (q Query) Run(seq int, args ...interface{}) error {
	acct := accts[seq]
	_, err := GetBalance(acct.Address, "xuper")
	return err
}

func (q Query) End(args ...interface{}) error {
	return nil
}
