package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuperchain/service/pb"
)

type evidence struct{
	total       int
	concurrency int

	generator   generate.Generator
	providers   []chan *pb.Transaction
}

func NewEvidence(config runner.Config) (Provider, error) {
	t := &evidence{
		total: int(config.N),
		concurrency: int(config.C),
	}
	if config.CEnd > config.C {
		t.concurrency = int(config.CEnd)
	}

	var err error
	conf := &generate.Config{
		Total: t.total,
		Concurrency: t.concurrency,
		Args: config.Tags,
	}
	t.generator, err = generate.NewEvidence(conf)
	if err != nil {
		return nil, fmt.Errorf("new evidence error: %v", err)
	}

	if err := t.generator.Init(); err != nil {
		return nil, fmt.Errorf("init generator error: %v", err)
	}

	t.providers = t.generator.Generate()
	return t, nil
}

func (t *evidence) DataProvider(run *runner.CallData) ([]*dynamic.Message, error) {
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
	RegisterProvider("evidence", NewEvidence)
}
