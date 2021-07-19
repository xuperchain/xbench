package generate

import (
	"fmt"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
)

type evidence struct {
	total       int
	concurrency int
	length      int
	batch       int

	accounts    []*account.Account
}

func NewEvidence(config *Config) (*evidence, error) {
	t := &evidence{
		total: config.Total,
		concurrency: config.Concurrency,
		batch: 1000,
	}

	var err error
	t.length, err = strconv.Atoi(config.Args["length"])
	if err != nil {
		return nil, fmt.Errorf("evidence length error: %v", err)
	}

	t.accounts, err = LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	log.Printf("generate: type=evidence, total=%d, concurrency=%d, length=%d",
		t.total, t.concurrency, t.length)
	return t, nil
}

func (t *evidence) Init() error {
	return nil
}

func (t *evidence) Generate() []chan *pb.Transaction {
	queues := make([]chan *pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan *pb.Transaction, t.concurrency)
	}

	var count int64
	total := t.total / t.concurrency
	provider := func(i int) {
		ak := t.accounts[i]
		for j := 0; j < total; j++ {
			tx := EvidenceTx(ak, t.length)
			queues[i] <- tx

			if (j+1) % t.concurrency == 0 {
				total := atomic.AddInt64(&count, int64(t.concurrency))
				if total%100000 == 0 {
					log.Printf("count=%d\n", total)
				}
			}
		}

		close(queues[i])
	}

	for i := 0; i < t.concurrency; i++ {
		go provider(i)
	}

	return queues
}

// 批量生成: 10w tps
// TODO: chan 读写成为瓶颈？
func (t *evidence) BatchGenerate() []chan []*pb.Transaction {
	queues := make([]chan []*pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan []*pb.Transaction, t.concurrency)
	}

	var count int64
	total := t.total / t.concurrency
	provider := func(i int) {
		ak := t.accounts[i]
		batch := make([]*pb.Transaction, t.batch)
		for j := 0; j < total; j++ {
			batch[j%t.batch] = EvidenceTx(ak, t.length)

			if (j+1) % t.batch == 0 {
				queues[i] <- batch
				batch = make([]*pb.Transaction, t.batch)

				total := atomic.AddInt64(&count, int64(t.batch))
				if total%100000 == 0 {
					log.Printf("count=%d\n", total)
				}
			}
		}

		close(queues[i])
	}

	for i := 0; i < t.concurrency; i++ {
		go provider(i)
	}

	return queues
}


func EvidenceTx(ak *account.Account, length int) *pb.Transaction {
	tx := &pb.Transaction{
		Version:   3,
		Desc:      RandBytes(length),
		Nonce:     strconv.FormatInt(time.Now().UnixNano(), 36),
		Timestamp: time.Now().UnixNano(),
		Initiator: ak.Address,
	}

	SignTx(tx, ak)
	return tx
}
