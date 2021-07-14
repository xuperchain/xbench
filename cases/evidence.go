package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xbench/generate"
	"strconv"
)

type evidence struct{
	total       int
	concurrency int
	batch       int

	length      int

	generator   generate.Generator
	provider    chan proto.Message
}

func NewEvidence(config runner.Config) (Provider, error) {
	length, err := strconv.Atoi(config.Tags["length"])
	t := &evidence{
		total: int(config.N),
		concurrency: int(config.C),
		batch: int(config.C),

		length: length,
	}

	t.generator, err = generate.NewEvidence(t.total, t.concurrency, t.length, t.batch)
	if err != nil {
		return nil, fmt.Errorf("new evidence error: %v", err)
	}

	t.provider = make(chan proto.Message, 10*t.concurrency)
	go func() {
		for txs := range t.generator.Generate() {
			for _, tx := range txs {
				t.provider <- tx
			}
		}
	}()
	return t, nil
}

func (t *evidence) DataProvider(*runner.CallData) ([]*dynamic.Message, error) {
	msg, ok := <- t.provider
	if !ok {
		return nil, fmt.Errorf("data provider close")
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