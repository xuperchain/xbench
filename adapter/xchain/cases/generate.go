package cases

import (
	"errors"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Generate struct {
	common.TestCase
}

// Init implements the comm.IcaseFace
func (g Generate) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.Connect(env.Host)
	amount := 0
	if env.Batch != 0 {
		amount = env.Batch
	} else {
		amount = env.Duration * 75000
	}
	Bank = lib.InitBankAcct("")
	for i:=0; i<parallel; i++ {
		Accts[i], _ = lib.CreateAcct()
	}
	//Accts = CreateTestClients(parallel, env.Host)
	log.INFO.Printf("prepare tokens of test accounts ...")
	for i := range Accts {
		rsp, err := lib.TransferSplit(Bank, Accts[i].Address, env.Chain, amount)
		if rsp.Header.Error != 0 || err != nil {
			log.ERROR.Printf("prepare tokens error: %#v, rsp: %#v", err, rsp)
			return errors.New("init token error")
		}
	}
	return nil
}

// Run implements the comm.IcaseFace
func (g Generate) Run(seq int, args ...interface{}) error {
	env := args[0].(common.TestEnv)
	rsp, err := lib.Transfer(Accts[seq], Bank.Address, env.Chain, "1")
	if rsp.Header.Error != 0 || err != nil {
		log.ERROR.Printf("transfer error: %#v, rsp: %#v", err, rsp)
		return errors.New("transfer error")
	}
	return nil
}

// End implements the comm.IcaseFace
func (g Generate) End(args ...interface{}) error {
	log.INFO.Printf("deal end")
	return nil
}
