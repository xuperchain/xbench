package generate

import "github.com/xuperchain/xuperchain/service/pb"

// generator
type Generator interface {
	// 业务初始化
	Init() error
	// 交易队列，单个队列内是关联交易，需要顺序发送；队列间是无关联交易，可以并发发送；
	Generate() []chan *pb.Transaction
}

// 配置文件
type Config struct {
	Host    string
	Total   int
	Concurrency int

	// 自定义参数
	Args    map[string]string
}

type NewGenerator func(config *Config) (Generator, error)
