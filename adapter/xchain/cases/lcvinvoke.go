package cases

import (
	"fmt"
	"time"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type LcvInvoke struct {
	common.TestCase
}

// Init implements the comm.IcaseFace
func (l LcvInvoke) Init(args ...interface{}) error {
	//parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	LBank = lib.GetAccountFromFile("", env.Chain, env.Host)
	LCBank = lib.CreateLcvContract(LBank, "1111111111111111", "proftestc")
	arg := map[string]string{"key": "key_0"}
	_, err := LCBank.Query("get", arg)
	if err != nil {
		delete(arg, "key")
		arg["creator"] = LBank.Acct.Address
		LCBank.Deploy(arg, "data/counter", "c")
	}
	return nil
}

// Run implements the comm.IcaseFace
func (l LcvInvoke) Run(seq int, args ...interface{}) error {
	k := fmt.Sprintf("key_%d", seq)
	arg := map[string]string{"key": k}
	_, err := LCBank.Invoke("increase", arg)
	// The xuperchain test-network with compilance check is not well-prepared yet.
	// Need to control the pressure.
	time.Sleep(time.Duration(1) * time.Second)
	return err
}

// End implements the comm.IcaseFace
func (l LcvInvoke) End(args ...interface{}) error {
	log.INFO.Printf("LCV invoke perf-test done.")
	return nil
}
