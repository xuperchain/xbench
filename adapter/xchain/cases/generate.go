package cases

import (
	"strconv"
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
	lib.SetCrypto(env.Crypto)
	amount := env.Batch
	Bank = lib.InitBankAcct("")
	addrs := []string{}
	for i:=0; i<parallel; i++ {
		Accts[i], _ = lib.CreateAcct(env.Crypto)
		addrs = append(addrs, Accts[i].Address)
		if len(Clis) < parallel {
			cli := lib.Conn(env.Host, env.Chain)
			Clis = append(Clis, cli)
		}
	}
	lib.InitIdentity(Bank, addrs, Clis[0])
	log.INFO.Printf("prepare tokens of test accounts ...")
	for i := range Accts {
		if env.Split {
			rsp, _, err := lib.Transplit(Bank, Accts[i].Address, amount, Clis[0])
			if rsp.Header.Error != 0 || err != nil {
				log.ERROR.Printf("prepare tokens error: %#v, rsp: %#v", err, rsp.Header)
				return errors.New("init token error")
			}
		} else {
			rsp, _, err := lib.Trans(Bank, Accts[i].Address, strconv.Itoa(amount), Clis[0])
			if rsp.Header.Error != 0 || err != nil {
				log.ERROR.Printf("prepare tokens error: %#v, rsp: %#v", err, rsp.Header)
				return errors.New("init token error")
			}
		}
	}
	return nil
}

// Run implements the comm.IcaseFace
func (g Generate) Run(seq int, args ...interface{}) error {
	rsp, _, err := lib.Trans(Accts[seq], Bank.Address, "1", Clis[seq])
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
