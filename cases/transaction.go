package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuperchain/service/pb"
)

type transaction struct{
	total       int
	concurrency int
	config      runner.Config

	providers   []chan *pb.Transaction
	generator   generate.Generator
}

func NewTransaction(config runner.Config) (Provider, error) {
	t := &transaction{
		total: int(config.N),
		concurrency: int(config.C),
		config: config,
	}

	if config.CEnd > config.C {
		t.concurrency = int(config.CEnd)
	}

	var err error
	conf := &generate.Config{
		Host: config.Host,
		Total: t.total,
		Concurrency: t.concurrency,
		Args: config.Tags,
	}
	t.generator, err = generate.NewTransfer(conf)
	if err != nil {
		return nil, fmt.Errorf("new generator error: %v", err)
	}

	if err := t.generator.Init(); err != nil {
		return nil, fmt.Errorf("init generator error: %v", err)
	}

	t.providers = t.generator.Generate()
	return t, nil
}

func (t *transaction) DataProvider(run *runner.CallData) ([]*dynamic.Message, error) {
	workID := generate.WorkID(run.WorkerID)
	tx, ok := <- t.providers[workID]
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