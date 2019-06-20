package demo

import (
	"errors"
	"math/rand"
	"time"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

// Deal 不像Nodejs等动态语言，我们无法直接在配置文件中定义每一个测试用例对应的回调函数
// 因此，这里使用了一点tricky的方法，即在配置文件里对每个测试用例定义Label标签，
// 然后将具体的压测区块链Label字段和测试用例Label字段进行比较，如果相同，表示是对应的回调函数！
type Deal struct {
	common.TestCase
}

// Init implements the comm.IcaseFace
func (d Deal) Init(args ...interface{}) error {
	log.INFO.Printf("deal init")
	return nil
}

// Run implements the comm.IcaseFace
func (d Deal) Run(seq int, args ...interface{}) error {
	base := int(8e4)
	t := rand.Intn(base)
	time.Sleep(time.Duration(t) * time.Nanosecond)
	if float64(t) > float64(base)*0.95 {
		return errors.New("timeout")
	}
	return nil
}

// End implements the comm.IcaseFace
func (d Deal) End(args ...interface{}) error {
	log.INFO.Printf("deal end")
	return nil
}
