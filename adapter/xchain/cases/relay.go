package cases

import (
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Relay struct {
	common.TestCase
}

func (r Relay) Init(args ...interface{}) error {
	return nil
}

func (r Relay) Run(seq int, args ...interface{}) error {
	return nil
}

func (r Relay) End(args ...interface{}) error {
	log.INFO.Printf("relay end")
	return nil
}
