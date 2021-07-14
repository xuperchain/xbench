package generate

import (
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
)

type evidence struct {
	// 总量
	total       int
	// 存证大小
	length      int
	// 并发量
	concurrency int
	// 一个批次生产的交易量
	batch       int

	queue       chan []*pb.Transaction
}

func NewEvidence(total, concurrency, length, batch int) (Generator, error) {
	t := &evidence{
		total: total,
		concurrency: concurrency,
		length: length,
		batch: batch,
		queue: make(chan []*pb.Transaction, 10*concurrency),
	}

	go func(t *evidence) {
		var count int64
		wg := new(sync.WaitGroup)
		for i := 0; i < t.concurrency; i++ {
			wg.Add(1)
			go func() {
				t.worker(&count)
				wg.Done()
			}()
		}
		wg.Wait()
		close(t.queue)
	}(t)
	return t, nil
}

func (t *evidence) Generate() chan []*pb.Transaction {
	return t.queue
}

func (t *evidence) worker(count *int64) {
	total := t.total / t.concurrency
	for i := 0; i < total; i += t.batch {
		txs := make([]*pb.Transaction, t.batch)
		for j := 0; j < t.batch; j++ {
			txs[j] = EvidenceTx(BankAK, t.length)
		}

		t.queue <- txs

		total := atomic.AddInt64(count, int64(t.batch))
		if total%100000 == 0 {
			log.Printf("count=%d\n", total)
		}
	}
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
