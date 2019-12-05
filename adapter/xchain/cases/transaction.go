package cases

import (
	"encoding/hex"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
)

type QueryTx struct {
	common.TestCase
}

var (
	qtxch []chan string
	last_txid string
)

// In this case, we run perfomance test with QueryTx.
// Iterate throuth refTx.

// Init implements the comm.IcaseFace
func (t QueryTx) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
	qtxch = make([]chan string, parallel)
	for i:=0; i<parallel; i++ {
		if len(Clis) < parallel {
			cli := lib.Conn(env.Host, env.Chain)
			Clis = append(Clis, cli)
		}
		qtxch[i] = make(chan string, 1)
	}
	rsp, _ := Clis[0].GetSystemStatus()
	for _, b := range(rsp.SystemsStatus.BcsStatus) {
		if b.Bcname == env.Chain {
			block := hex.EncodeToString(b.UtxoMeta.LatestBlockid)
			binfo, _ := Clis[0].GetBlock(block)
			last_txid = hex.EncodeToString(binfo.Block.Transactions[0].Txid)
			for i:=0; i<parallel; i++ {
				qtxch[i] <- last_txid
			}
			break
		}
	}
	return nil
}

// Run implements the comm.IcaseFace
func (t QueryTx) Run(seq int, args ...interface{}) error {
	txid := <-qtxch[seq]
	rsp, err := Clis[seq].QueryTx(txid)
	reftx := rsp.Tx.TxInputs
	if len(reftx) > 0 {
		qtxch[seq] <- hex.EncodeToString(reftx[0].RefTxid)
	} else {
		qtxch[seq] <- last_txid
	}
	return err
}

// End implements the comm.IcaseFace
func (t QueryTx) End(args ...interface{}) error {
	log.INFO.Printf("QueryTx perf-test done.")
	return nil
}
