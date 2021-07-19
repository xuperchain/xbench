package generate

import (
	"context"
	"fmt"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"sync"
	"sync/atomic"
)

const (
	Benchmark = "benchmark"
	BenchmarkTransfer = "transfer"
	BenchmarkTransaction = "transaction"
	BenchmarkEvidence = "evidence"
	BenchmarkContract = "contract"
	BenchmarkFile = "file"
)

// generator
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

// 生产交易
func Producer(ctx context.Context, generator Generator, queues []chan *pb.Transaction) {
	var total int64
	length := len(queues)
	wg := new(sync.WaitGroup)
	wg.Add(length)
	for i := 0; i < length; i++ {
		go func(i int) {
			defer wg.Done()
			var count int64
			for {
				tx, err := generator.Generate(i)
				if err != nil {
					log.Fatalf("generate tx error: %v", err)
					return
				}

				select {
				case <-ctx.Done():
					return
				case queues[i] <- tx:
				}

				count++
				if (count) % 1000 == 0 {
					if atomic.AddInt64(&total, count) % 100000 == 0 {
						log.Printf("total=%d", total)
					}
					count = 0
				}
			}
		}(i)
	}
	wg.Done()
}

// 消费交易
func Consumer(ctx context.Context, queues []chan *pb.Transaction, consume func(i int, tx *pb.Transaction) error, total int) {
	wg := new(sync.WaitGroup)
	wg.Add(len(queues))
	for i, queue := range queues {
		go func(i int, queue chan *pb.Transaction) {
			defer wg.Done()
			count := 0
			for {
				select {
				case <-ctx.Done():
					return
				case tx, ok := <- queue:
					if !ok {
						return
					}

					if err := consume(i, tx); err != nil {
						return
					}
				}

				count++
				if total > 0 && count >= total {
					return
				}
			}
		}(i, queue)
	}
	wg.Wait()
}
