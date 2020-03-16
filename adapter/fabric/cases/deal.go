package cases

import (
	"fmt"
	"github.com/xuperchain/xuperbench/adapter/fabric/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Deal struct {
	common.TestCase
}

// Init implements the comm.IcaseFace
func (d Deal) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	sdk, _ := lib.InitSDK(SDK_CONF)
	for i:=0; i<parallel; i++ {
		cli, _ := lib.CreateNode(env.Chain, USER, env.Host, sdk)
		Clis = append(Clis, cli)
	}
	log.INFO.Printf("prepare connections ...")
	return nil
}

// Run implements the comm.IcaseFace
func (d Deal) Run(seq int, args ...interface{}) error {
	key := fmt.Sprintf("key_%d", seq)
	param := []string{key}
	_, err := Clis[seq].Execute(CHAINCODE, "increase", param)
	return err
}

// End implements the comm.IcaseFace
func (d Deal) End(args ...interface{}) error {
	return nil
}
