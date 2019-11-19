package cases

import (
	"time"
	"strconv"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
)

type LcvTrans struct {
	common.TestCase
}

// Init implements the comm.IcaseFace
func (l LcvTrans) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	amount := strconv.Itoa(env.Batch * 200)
	LBank = lib.GetAccountFromFile("", env.Chain, env.Host)
	Cfg := lib.GenLcvConfig(env.Endorse, env.XCheck, LBank.Acct.Address)
	LBank.SetCfg(Cfg)
	LCBank = lib.CreateLcvContract(LBank, "1111111111111111", "unified_check")
	LCBank.SetCfg(Cfg)
	addrs := []string{}
	lib.BatchRetrieve(LAccts, parallel, env.Chain, env.Host)
	for i:=0; i<parallel; i++ {
		LAccts[i].SetCfg(Cfg)
		LBank.Transfer(LAccts[i].Acct.Address, amount, "0", "")
		addrs = append(addrs, LAccts[i].Acct.Address)
	}
	lib.LcvInitIdentity(addrs, LCBank)
	return nil
}

// Run implements the comm.IcaseFace
func (l LcvTrans) Run(seq int, args ...interface{}) error {
	_, err := LAccts[seq].Transfer(LBank.Acct.Address, "100", "0", "")
	time.Sleep(time.Duration(1) * time.Second)
	return err
}

// End implements the comm.IcaseFace
func (l LcvTrans) End(args ...interface{}) error {
	return nil
}
