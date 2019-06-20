package common

import (
	"errors"
	"fmt"
)

type TestCase struct {
	BCType BlockChain
	Label  CaseType
}

func (s TestCase) Init(args ...interface{}) error {
	return errors.New("init Not implemented")
}

func (s TestCase) Run(seq int, args ...interface{}) error {
	return errors.New("run Not implemented")
}

func (s TestCase) End(args ...interface{}) error {
	return errors.New("end Not implemented")
}

func (s TestCase) String() string {
	return fmt.Sprintf("%v-%v", string(s.BCType), string(s.Label))
}

func (s TestCase) GetTestCase() TestCase {
	return s
}

type Getter interface {
	GetTestCase() TestCase
}

type TestEnv struct {
	Host string
	Batch int
	Duration int
	Chain string
}

// BenchMsg contains all the bench info while doing one bench test.
type BenchMsg struct {
	// 压测的区块链类型
	TestCase

	// 压测轮数
	TxNumber int

	// 压测时间(单位：s)，如果 TxNumber 和 Duration 两个值同时设置了，则以 Duration 设置为准
	TxDuration int

	// 并发数
	Parallel int

	// 每一个压测测试用例具体对应的回调函数
	CB ICaseFace

	Env TestEnv

	// 回传给CB回调函数的参数
	// 用法： CB(Args)
	Args []interface{}
}

type BehavMsg struct {
	TestCase
	CB ICaseFace
	Args []interface{}
}

// ICaseFace 压测接口，每一个适配区块链的每一个测试用例，都需要实现这三个接口！
type ICaseFace interface {
	Init(args ...interface{}) error
	Run(seq int, args ...interface{}) error
	End(args ...interface{}) error

	Getter
}
