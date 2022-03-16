package cases

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

const (
	Benchmark = "benchmark"
	// 转账：调用sdk生成转账数据
	CaseTransfer = "transfer"
	// 转账: 离线生成数据，没有进行 SelectUTXO
	CaseTransaction = "transaction"
	// 存证
	CaseEvidence = "evidence"
	// 合约
	CaseContract = "contract"
	// 文件：json格式的tx数据
	CaseFile = "file"
)

// 生成压测用例接口
type Generator interface {
	// 业务初始化
	Init() error
	// 根据并发ID获取构造的交易数据
	Generate(id int) (proto.Message, error)
}

// 配置文件
type Config struct {
	Host        string
	Total       int
	Concurrency int
	// 自定义参数
	Args map[string]string
}

type NewGenerator func(config *Config) (Generator, error)

var generators = make(map[string]NewGenerator, 8)

func RegisterGenerator(name string, contract NewGenerator) {
	generators[name] = contract
}

func GetGenerator(name string, config *Config) (Generator, error) {
	if newContract, ok := generators[name]; ok {
		return newContract(config)
	}

	return nil, fmt.Errorf("generator not exist")
}
