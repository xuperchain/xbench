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
		tx := lib.ProfTx(Accts[i], Bank.Address, Clis[i])
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
	lib.SetCrypto(env.Crypto)
	amount := env.Batch
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
	txstore = make([]ch, parallel)
	wg.Add(parallel)
	for i, _ := range txstore {
		txstore[i] = make(chan *pb.TxStatus, amount)
	}
	log.INFO.Printf("prepare tokens of test accounts ...")
	txid := ""
	for i := range Accts {
		rsp, x, err := lib.Transplit(Bank, Accts[i].Address, amount, Clis[0])
		txid = x
		if rsp.Header.Error != 0 || err != nil {
			log.ERROR.Printf("init token error: %#v", err)
			return errors.New("init token error")
		}
	}
	log.INFO.Printf("prepere tx of test accounts ...")
	lib.WaitConfirm(txid, 5, Clis[0])
	for k := range Accts {
		go createtx(k, amount, env.Chain)
	}
	wg.Wait()
	return nil
}

func (d Deal) Run(seq int, args ...interface{}) error {
	txs := <-txstore[seq]
	rsp, _, err := Clis[seq].PostTx(txs)
	if rsp.Header.Error != 0 || err != nil {
		return errors.New("run posttx error")
	}
	return nil
}

func (d Deal) End(args ...interface{}) error {
	return nil
}
