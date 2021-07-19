package cases

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/xuperchain/xuperchain/service/pb"
)

const (
	Benchmark = "benchmark"
	// 转账：调用sdk生成转账数据: SelectUTXO
	CaseTransfer = "transfer"
	// 转账: not SelectUTXO
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
	Generate(id int) (*pb.Transaction, error)
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


type Consume func(i int, tx *pb.Transaction) error
func Consumer(total, concurrency int, generator Generator, consume Consume) {
	var inc int64
	wg := new(sync.WaitGroup)
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			var count int
			for {
				tx, err := generator.Generate(i)
				if err != nil {
					log.Fatalf("generate tx error: %v", err)
					return
				}

				if err = consume(i, tx); err != nil {
					return
				}

				count++

				if count % 1000 == 0 {
					newInc := atomic.AddInt64(&inc, 1000)
					if newInc % 100000 == 0 {
						log.Printf("total=%d", newInc)
					}
				}

				if count >= total {
					return
				}
			}
		}(i)
	}
	wg.Wait()
}
