package xchain

import (
	"errors"
	"sync"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperunion/pb"
)

type Deal struct {
	common.TestCase
}

type ch chan *pb.TxStatus

var (
	txstore = []ch{}
	wg = sync.WaitGroup{}
)

func createtx(i int, batch int, chain string) {
	for c:=0; c<batch; c++ {
		tx := GenProfTx(Accts[i], Bank.Address, chain)
		if i == 0 && c > 0 && c % 500 == 0 {
			log.DEBUG.Printf("gen %d Tx", c)
		}
		txstore[i] <- tx
	}
	wg.Done()
}

func (d Deal) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	Connect(env.Host)
	amount := 0
	if env.Batch != 0 {
		amount = env.Batch
	} else {
		amount = env.Duration * 75000
	}
	Bank = InitBankAcct()
	Accts = CreateTestClients(parallel, env.Host)
	txstore = make([]ch, parallel)
	wg.Add(parallel)
	for i, _ := range txstore {
		txstore[i] = make(chan *pb.TxStatus, amount)
	}
	log.INFO.Printf("prepare tokens of test accounts ...")
	for i := range Accts {
		rsp, err := TransferSplit(Bank, Accts[i].Address, env.Chain, amount)
		if rsp.Header.Error != 0 || err != nil {
			log.ERROR.Printf("init token error: %#v", err)
			return errors.New("init token error")
		}
	}
	log.INFO.Printf("prepere tx of test accounts ...")
	for k := range Accts {
		go createtx(k, amount, env.Chain)
	}
	wg.Wait()
	return nil
}

func (d Deal) Run(seq int, args ...interface{}) error {
	tx := <-txstore[seq]
	rsp, err := PostTx(tx)
	if rsp.Header.Error != 0 || err != nil {
		return errors.New("run posttx error")
	}
	return nil
}

func (d Deal) End(args ...interface{}) error {
	return nil
}
