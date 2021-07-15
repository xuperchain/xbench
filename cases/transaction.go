package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"strconv"
	"strings"
)

type transaction struct{
	total       int
	split       int
	concurrency int

	client      *xuper.XClient
	accounts    []*account.Account
	providers   []chan *pb.Transaction
	generator   generate.SDKGenerator
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
	t.generator, err = generate.NewTransfer(config.Host, t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("new generator error: %v", err)
	}

	if err := t.generator.Init(); err != nil {
		return nil, fmt.Errorf("init generator error: %v", err)
	}

	t.providers = t.generator.Generate()
	return t, nil
}

func getWorkID(workID string) int {
	workIdStr := strings.Split(workID[1:], "c")[0]
	workId, _ := strconv.Atoi(workIdStr)
	return workId
}

func (t *transaction) DataProvider(run *runner.CallData) ([]*dynamic.Message, error) {
	workID := getWorkID(run.WorkerID)
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