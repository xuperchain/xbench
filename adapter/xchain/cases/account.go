package cases

import (
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type QueryAcct struct {
	common.TestCase
}

// In this case, we run perfomance test with GetBalance rpc request

// Init implements the comm.IcaseFace
func (a QueryAcct) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
	Bank = lib.InitBankAcct("")
	Accts[0], _ = lib.CreateAcct(env.Crypto)
	for i:=0; i<parallel; i++ {
		if len(Clis) < parallel {
			cli := lib.Conn(env.Host, env.Chain)
			Clis = append(Clis, cli)
		}
	}
	lib.Trans(Bank, Accts[0].Address, "12345", Clis[0])
	return nil
}

// Run implements the comm.IcaseFace
func (a QueryAcct) Run(seq int, args ...interface{}) error {
	rsp, err := Clis[seq].GetBalance(Accts[0].Address)
	if rsp == nil || len(rsp.Bcs) == 0 {
		log.ERROR.Printf("Query account error. rsp: %#v, err: %#v", rsp, err)
	}
	return err
}

// End implements the comm.IcaseFace
func (a QueryAcct) End(args ...interface{}) error {
	log.INFO.Printf("Query account end.")
	return nil
}
