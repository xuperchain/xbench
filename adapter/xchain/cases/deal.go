package cases

import (
	"errors"
	"sync"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
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
		tx := lib.GenProfTx(Accts[i], Bank.Address, chain)
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
	lib.Connect(env.Host, env.Nodes, env.Crypto)
	amount := 0
	if env.Batch != 0 {
		amount = env.Batch
	} else {
		amount = env.Duration * 75000
	}
	Bank = lib.InitBankAcct("")
	addrs := []string{}
	for i:=0; i<parallel; i++ {
		Accts[i], _ = lib.CreateAcct(env.Crypto)
		addrs = append(addrs, Accts[i].Address)
	}
	lib.InitIdentity(Bank, env.Chain, addrs)
	//Accts = CreateTestClients(parallel, env.Host)
	txstore = make([]ch, parallel)
	wg.Add(parallel)
	for i, _ := range txstore {
		txstore[i] = make(chan *pb.TxStatus, amount)
	}
	lasttx := ""
	log.INFO.Printf("prepare tokens of test accounts ...")
	for i := range Accts {
		rsp, txid, err := lib.TransferSplit(Bank, Accts[i].Address, env.Chain, amount)
		if rsp.Header.Error != 0 || err != nil {
			log.ERROR.Printf("init token error: %#v", err)
			return errors.New("init token error")
		}
		lasttx = txid
	}
	// wait
	if !lib.WaitTx(10, lasttx, env.Chain) {
		return errors.New("wait timeout")
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
	rsp, err := lib.PostTx(tx)
	if rsp.Header.Error != 0 || err != nil {
		return errors.New("run posttx error")
	}
	return nil
}

func (d Deal) End(args ...interface{}) error {
	return nil
}
