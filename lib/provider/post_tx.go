package provider

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xbench/cases"
	"github.com/xuperchain/xbench/lib"
	"github.com/xuperchain/xuperchain/service/pb"
)

type postTx struct{
	benchmark   string
	concurrency int
	config      *runner.Config

	generator   cases.Generator
}

func NewPostTx(config *runner.Config) (Provider, error) {
	t := &postTx{
		benchmark: config.Tags[cases.Benchmark],
		concurrency: int(config.C),
		config: config,
	}

	if config.CEnd > config.C {
		t.concurrency = int(config.CEnd)
	}

	var err error
	conf := &cases.Config{
		Host: config.Host,
		Concurrency: t.concurrency,
		Args: config.Tags,
	}
	t.generator, err = cases.GetGenerator(t.benchmark, conf)
	if err != nil {
		return nil, fmt.Errorf("new generator error: %v, benchmark=%s", err, t.benchmark)
	}

	if err := t.generator.Init(); err != nil {
		return nil, fmt.Errorf("init generator error: %v, benchmark=%s", err, t.benchmark)
	}

	return t, nil
}

func (t *postTx) DataProvider(run *runner.CallData) ([]*dynamic.Message, error) {
	workID := lib.WorkID(run.WorkerID)
	tx, err :=  t.generator.Generate(workID)
	if err != nil {
		return nil, err
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
	RegisterProvider(CallPostTx, NewPostTx)
}
