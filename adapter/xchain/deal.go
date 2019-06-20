package xchain

import (
	"time"
	"strconv"
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
	testbank = &Acct{}
	testaccts = map[int]*Acct{}
	txstore = []ch{}
	wg = sync.WaitGroup{}
)

func createtx(i int, batch int, chain string) {
	for c:=0; c<batch; c++ {
		tx, err := GenTx(testaccts[i], testbank.Address, chain, "1", "", "0")
		if err != nil {
			log.ERROR.Printf("genTx error: %#v", err)
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
	testbank = InitBankAcct()
	testaccts = CreateTestClients(parallel, env.Host)
	txstore = make([]ch, parallel)
	wg.Add(parallel)
	for i, _ := range txstore {
		txstore[i] = make(chan *pb.TxStatus, amount)
	}
	for i := range testaccts {
		t, err := GenTx(testbank, testaccts[i].Address, env.Chain, strconv.Itoa(amount), "", "0")
		if err != nil {
			log.ERROR.Printf("init GenTx error: %#v", err)
			return err
		}
		PostTx(t)
	}
	log.INFO.Printf("prepare tokens of test accounts ...")
	time.Sleep(4 * time.Second)
	for j := range testaccts {
		SplitUTXO(testaccts[j], env.Chain, amount)
	}
	log.INFO.Printf("prepare utxos of test accounts ...")
	time.Sleep(4 * time.Second)
	for k := range testaccts {
		go createtx(k, amount, env.Chain)
	}
	wg.Wait()
	log.INFO.Printf("prepere tx of test accounts ...")
	return nil
}

func (d Deal) Run(seq int, args ...interface{}) error {
	tx := <-txstore[seq]
	_, err := PostTx(tx)
	return err
}

func (d Deal) End(args ...interface{}) error {
	return nil
}
