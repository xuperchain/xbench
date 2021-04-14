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
	qacct         = "1123581321345589"
	qcontract     = "counter"
	qcontractpath = "data/counter"
)

// In this case, we run perfomance test with contract query.

// Init implements the comm.IcaseFace
func (q Query) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
	for i := 0; i <= parallel-1 && len(Clis) < parallel; i++ {
		cli := lib.Conn(env.Host, env.Chain)
		Clis = append(Clis, cli)
	}
	Bank = lib.InitBankAcct("")
	log.INFO.Printf("check contract account ...")
	account := fmt.Sprintf("XC%s@%s", qacct, env.Chain)
	status, err := Clis[0].QueryACL(account)
	if !status.Confirmed {
		// lib.NewContractAcct(Bank, qacct, Clis[0])
	}
	log.INFO.Printf("check counter contract ...")
	_, _, err = lib.QueryContract(Bank, contract, "Get", "key_0", Clis[0])
	if err != nil {
		lib.Trans(Bank, account, "10000000", Clis[0])
		// lib.DeployContract(Bank, contractpath, account, qcontract, Clis[0])
	}
	lib.InvokeContract(Bank, qcontract, "increase", "key_0", Clis[0])
	log.INFO.Printf("prepare done %s on %s", account, qcontract)
	return nil
}

// Run implements the comm.IcaseFace
func (q Query) Run(seq int, args ...interface{}) error {
	rsp, _, err := lib.QueryContract(Bank, qcontract, "Get", "key_0", Clis[seq])
	if rsp == nil || err != nil {
		return errors.New("run query error")
	}
	return err
}

// End implements the comm.IcaseFace
func (q Query) End(args ...interface{}) error {
	return nil
}
