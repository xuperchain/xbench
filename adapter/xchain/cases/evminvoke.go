package cases

import (
	"errors"
	"fmt"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type EVMInvoke struct {
	common.TestCase
}

var (
	evmAcct        = "1123581321345589"
	evmContract    = "proftestc"
	evmContractAbi = "data/Counter.abi"
	evmContractBin = "data/Counter.bin"
)

// Init implements the comm.IcaseFace
func (i EVMInvoke) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
	txid := ""
	for i := 0; i <= parallel-1 && len(Clis) < parallel; i++ {
		cli := lib.Conn(env.Host, env.Chain)
		Clis = append(Clis, cli)
	}
	Bank = lib.InitBankAcct("")
	log.INFO.Printf("check evm contract account ...")
	account := fmt.Sprintf("XC%s@%s", evmAcct, env.Chain)
	status, err := Clis[0].QueryACL(account)
	if err != nil || status == nil {
		log.ERROR.Printf("queryacl error")
		return err
	}
	if !status.Confirmed {
		log.INFO.Printf("new evm contract account ...")
		_, txid, _ = lib.NewContractAcct(Bank, evmAcct, Clis[0])
		lib.WaitConfirm(txid, 5, Clis[0])
	}
	log.INFO.Printf("check evm counter contract ...")
	_, _, err = lib.QueryEVMContract(Bank, evmContractAbi, evmContract, "get", "key_0", Clis[0])
	if err != nil {
		_, txid, _ = lib.Trans(Bank, account, "10000000", Clis[0])
		lib.WaitConfirm(txid, 5, Clis[0])
		_, txid, _ = lib.DeployEVMContract(Bank, evmContractBin, evmContractAbi, account, evmContract, Clis[0])
		lib.WaitConfirm(txid, 5, Clis[0])
	}
	log.INFO.Printf("prepare done %s on %s", account, evmContract)
	return nil
}

func (i EVMInvoke) Run(seq int, args ...interface{}) error {
	k := fmt.Sprintf("key_%d", seq)
	rsp, _, err := lib.InvokeEVMContract(Bank, evmContractAbi, evmContract, "increase", k, Clis[seq])
	if err != nil || rsp.Header.Error != 0 {
		log.ERROR.Printf("err on invoke %#v", rsp.Header)
		return errors.New("invoke evm contract error")
	}
	return nil
}

func (i EVMInvoke) End(args ...interface{}) error {
	return nil
}
