package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"math/rand"
	"sync"
	"time"
)

type transaction struct{
	total       int
	split       int
	concurrency int

	client      *xuper.XClient
	accounts    []*account.Account
	provider    chan *pb.Transaction
}

func NewTransaction(config runner.Config) (Provider, error) {
	t := &transaction{
		total: int(config.N),
		concurrency: int(config.C),
	}

	if config.CEnd > config.C {
		t.concurrency = int(config.CEnd)
	}

	var err error
	t.accounts, err = generate.LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	t.client, err = xuper.New(config.Host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}

	err = generate.Transfer(t.client, generate.BankAK, t.accounts, "100000000", 10)
	if err != nil {
		return nil, fmt.Errorf("transfer to test accounts error: %v", err)
	}

	go t.generateTx()
	return t, nil
}

func (t *transaction) generateTx() {
	t.provider = make(chan *pb.Transaction, 2*t.concurrency)

	length := len(t.accounts)
	loop := make(chan *account.Account, length)
	for _, ak := range t.accounts {
		loop <- ak
	}

	wg := new(sync.WaitGroup)
	wg.Add(t.concurrency)
	provider := func() {
		defer wg.Done()
		for from := range loop {
			to := t.accounts[rand.Intn(length)]
			tx, err := t.client.Transfer(from, to.Address, "10", xuper.WithNotPost())
			if err != nil {
				log.Printf("generate tx error: %v, address=%s", err, from.Address)
				time.Sleep(100*time.Millisecond)
				loop <- from
				continue
			}

			t.provider <- tx.Tx
			loop <- from
		}
	}

	for i := 0; i < t.concurrency; i++ {
		go provider()
	}

	wg.Wait()
	close(t.provider)
}

func (t *transaction) DataProvider(*runner.CallData) ([]*dynamic.Message, error) {
	tx, ok := <- t.provider
	if !ok {
		return nil, fmt.Errorf("data provider close")
	}

	msg := &pb.TxStatus{
		Bcname: BlockChain,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Txid:   tx.Txid,
	}
	dynamicMsg, err := dynamic.AsDynamicMessage(msg)
	if err != nil {
		return nil, err
	}

	return []*dynamic.Message{dynamicMsg}, nil
}

func init() {
	RegisterProvider("transaction", NewTransaction)
}