package cases

import (
	"sync"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Relay struct {
	common.TestCase
}

var (
	relay = []string{}
	rg = sync.WaitGroup{}
)

func firstlap(i int) {
	_, txid, _ := lib.Trans(Bank, Accts[i].Address, "1", Clis[i])
	relay[i] = txid
	lib.WaitConfirm(txid, 5, Clis[i])
	rg.Done()
}

// In this case, we run perfomance test with Transaction (without selectUTXO).

// Init implements the comm.IcaseFace
func (r Relay) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
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
	relay = make([]string, parallel, parallel)
	rg.Add(parallel)
	for i:=0; i<parallel; i++ {
		go firstlap(i)
	}
	rg.Wait()
	log.INFO.Printf("prepare done.")
	return nil
}

// Run implements the comm.IcaseFace
func (r Relay) Run(seq int, args ...interface{}) error {
	tx := lib.FormatTx(Accts[seq].Address)
	lib.FormatOutput(tx, Accts[seq].Address, "1", "0")
	//rsp, _, _ := Clis[seq].PreExec(nil, "", "", "", Accts[seq].Address)
	lib.FormatRelayInput(tx, relay[seq], nil)
	txs := Clis[seq].SignTx(tx, Accts[seq], "")
	_, txid, err := Clis[seq].PostTx(txs)
	relay[seq] = txid
	return err
}

// End implements the comm.IcaseFace
func (r Relay) End(args ...interface{}) error {
	log.INFO.Printf("Relay perf-test done.")
	return nil
}
