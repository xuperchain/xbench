package cases

import (
	"encoding/hex"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type QueryBlock struct {
	common.TestCase
}

var (
	blockch []chan string
	last_block string
)

// In this case, we run perfomance test with GetBlock rpc request
// iterate the chain from the last block to the gensis block.

// Init implements the comm.IcaseFace
func (b QueryBlock) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
	blockch = make([]chan string, parallel)
	for i:=0; i<parallel; i++ {
		if len(Clis) < parallel {
			cli := lib.Conn(env.Host, env.Chain)
			Clis = append(Clis, cli)
		}
		blockch[i] = make(chan string, 1)
	}
	rsp, _ := Clis[0].GetSystemStatus()
	for _, b := range(rsp.SystemsStatus.BcsStatus) {
		if b.Bcname == env.Chain {
			last_block = hex.EncodeToString(b.UtxoMeta.LatestBlockid)
			for i:=0; i<parallel; i++ {
				blockch[i] <- last_block
			}
			break
		}
	}
	return nil
}

// Run implements the comm.IcaseFace
func (b QueryBlock) Run(seq int, args ...interface{}) error {
	blockid := <-blockch[seq]
	rsp, err := Clis[seq].GetBlock(blockid)
	if err != nil {
		return err
	}
	preid := hex.EncodeToString(rsp.Block.PreHash)
	if preid != "" {
		blockch[seq] <- preid
	} else {
		blockch[seq] <- last_block
	}
	return err
}

// End implements the comm.IcaseFace
func (b QueryBlock) End(args ...interface{}) error {
	log.INFO.Printf("Query block end.")
	return nil
}
