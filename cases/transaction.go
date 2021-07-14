package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
)

type transaction struct{
	total       int
	split       int
	concurrency int

	txid string
	address string
	amount string

	generator   generate.Generator
	provider    chan *pb.Transaction
}

func NewTransaction(config runner.Config) (Provider, error) {
	t := &transaction{
		total: int(config.N),
		concurrency: int(config.C),
		split: int(config.C),

		txid: config.Tags["txid"],
		address: config.Tags["address"],
		amount: config.Tags["amount"],
	}

	accounts, err := generate.LoadAccount(t.split)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	bootTx, err := generate.BootTx(t.txid, t.address, t.amount)
	if err != nil {
		return nil, err
	}

	log.Printf("bootTx=%s", generate.FormatTx(bootTx))

	t.generator, err = generate.NewTransaction(t.total, t.concurrency, t.split, accounts, bootTx)
	if err != nil {
		return nil, fmt.Errorf("new evidence error: %v", err)
	}

	t.provider = make(chan *pb.Transaction, t.concurrency)
	go func() {
		for txs := range t.generator.Generate() {
			for _, tx := range txs {
				t.provider <- tx
			}
		}
	}()
	return t, nil
}

func (t *transaction) DataProvider(*runner.CallData) ([]*dynamic.Message, error) {
	tx, ok := <- t.provider
	if !ok {
		return nil, fmt.Errorf("data provider close")
	}

	//log.Printf("%s", generate.FormatTx(tx))
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