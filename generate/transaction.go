package generate

import (
	"context"
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/account"
	pb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	"github.com/xuperchain/xupercore/protos"
	"log"
	"math"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type transaction struct {
	total int
	split int
	concurrency int

	ctx context.Context

	accounts    []*account.Account
	addresses   map[string]*account.Account

	level       int
	levelCh     chan Level

	queue       chan []*pb.Transaction
}

type Level struct {
	txs []*pb.Transaction
	level int
}

func NewTransaction(total, concurrency, split int, accounts []*account.Account, bootTx *pb.Transaction) (Generator, error) {
	if total <= 1 {
		return nil, fmt.Errorf("split error: split=%v", total)
	}

	if len(accounts) < split {
		return nil, fmt.Errorf("accounts not enough: split=%d, accounts=%d", split, len(accounts))
	}

	addresses := make(map[string]*account.Account, len(accounts))
	for _, ak := range accounts {
		addresses[ak.Address] = ak
	}

	t := &transaction{
		total: total,
		concurrency: concurrency,
		accounts: accounts,
		addresses: addresses,
		levelCh: make(chan Level, 10*split),
		queue: make(chan []*pb.Transaction, 10*split),
	}

	level := math.Log10(float64(total)) / math.Log10(float64(split))
	t.level = int(math.Ceil(level))
	log.Printf("transaction: total=%d split=%d level=%d", total, split, t.level)

	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx

	go t.output(cancel)

	go func() {
		t.generate(bootTx, 0)
		select {
		case <-t.ctx.Done():
			return
		}
	}()

	return t, nil
}

func (t *transaction) Generate() chan []*pb.Transaction {
	return t.queue
}

// 生产交易
func (t *transaction) generate(tx *pb.Transaction, level int) {
	if level >= t.level {
		return
	}

	select {
	case <-t.ctx.Done():
		return
	default:
	}

	unconfirmedTxs := t.producer(tx, t.split)
	t.levelCh <- Level{
		txs: unconfirmedTxs,
		level: level,
	}
	for _, unconfirmedTx := range unconfirmedTxs {
		t.generate(unconfirmedTx, level+1)
	}
}

// 产出交易：level之间是关联交易，level内是无关联交易
func (t *transaction) output(cancel context.CancelFunc) {
	var count int64
	for {
		select {
		case txsLevel := <-t.levelCh:
			//t.store(txsLevel.txs, txsLevel.level)
			t.queue <- txsLevel.txs
			if txsLevel.level+1 < t.level {
				continue
			}

			ch := make(chan struct{}, t.concurrency)
			for i := 0; i < t.concurrency; i++ {
				ch <- struct{}{}
			}

			wg := new(sync.WaitGroup)
			for _, tx := range txsLevel.txs {
				wg.Add(1)
				<- ch
				go func(tx *pb.Transaction) {
					txs := t.producer(tx, 1)
					//t.store(txs, txsLevel.level+1)
					t.queue <- txs

					total := atomic.AddInt64(&count, int64(len(txs)))
					if total%100000 == 0 {
						log.Printf("count=%d\n", count)
					}

					ch <- struct {}{}
					wg.Done()
				}(tx)

			}
			wg.Wait()

			if count >= int64(t.total) {
				cancel()
				return
			}
		}
	}
}

// 生成子交易
func (t *transaction) producer(tx *pb.Transaction, split int) []*pb.Transaction {
	txs := make([]*pb.Transaction, len(tx.TxOutputs))
	for i, txOutput := range tx.TxOutputs {
		if txOutput == nil {
			continue
		}

		ak := t.addresses[string(txOutput.ToAddr)]
		if ak == nil {
			ak = AK
		}

		input := &protos.TxInput{
			RefTxid: tx.Txid,
			RefOffset: int32(i),
			FromAddr: txOutput.ToAddr,
			Amount: txOutput.Amount,
		}
		output := &protos.TxOutput{
			ToAddr: []byte(t.accounts[i].Address),
			Amount: txOutput.Amount,
		}
		newTx := TransferTx(input, output, split)
		txs[i] = SignTx(newTx, ak)
	}

	return txs
}

func TransferTx(input *protos.TxInput, output *protos.TxOutput, split int) *pb.Transaction {
	tx := &pb.Transaction{
		Version:   3,
		//Desc:      RandBytes(200),
		Nonce:     strconv.FormatInt(time.Now().UnixNano(), 36),
		Timestamp: time.Now().UnixNano(),
		Initiator: string(input.FromAddr),
		TxInputs: []*protos.TxInput{input},
		TxOutputs: Split(output, split),
	}

	return tx
}

func Split(txOutput *protos.TxOutput, split int) []*protos.TxOutput {
	if split <= 1 {
		return []*protos.TxOutput{txOutput}
	}

	total := big.NewInt(0).SetBytes(txOutput.Amount)
	if big.NewInt(int64(split)).Cmp(total) == 1 {
		// log.Printf("split utxo <= balance required")
		panic("amount not enough")
		return []*protos.TxOutput{txOutput}
	}

	amount := big.NewInt(0)
	amount.Div(total, big.NewInt(int64(split)))

	output := protos.TxOutput{}
	output.Amount = amount.Bytes()
	output.ToAddr = txOutput.ToAddr

	rest := total
	txOutputs := make([]*protos.TxOutput, 0, split+1)
	for i := 1; i < split && rest.Cmp(amount) == 1; i++ {
		tmpOutput := output
		txOutputs = append(txOutputs, &tmpOutput)
		rest.Sub(rest, amount)
	}
	output.Amount = rest.Bytes()
	txOutputs = append(txOutputs, &output)

	return txOutputs
}
