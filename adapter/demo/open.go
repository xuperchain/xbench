package demo

import (
	"errors"
	"math/rand"
	"time"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

// Open ...
type Open struct {
	common.TestCase
}

// Init ...
func (d Open) Init(args ...interface{}) error {
	log.INFO.Printf("open init")
	return nil
}

// Run ...
func (d Open) Run(seq int, args ...interface{}) error {
	base := int(3e4)
	t := rand.Intn(base)
	time.Sleep(time.Duration(t) * time.Nanosecond)
	if float64(t) > float64(base)*0.9 {
		return errors.New("timeout")
	}
	return nil
}

// End ...
func (d Open) End(args ...interface{}) error {
	log.INFO.Printf("open end")
	return nil
}
