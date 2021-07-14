package generate

import (
	"context"
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/lib/utils"
	"log"
	"math"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// 离线生成交易算法描述：递归的分裂交易，直到达到目标数量
// total = split^level
// total: 生成交易总数(是split的整数倍)
// split: 分裂系数，表示一个交易分裂出split个子交易 => len(txOutput)=split
// level: 分裂次数，第一次分裂出的交易level=0，不同level之间是关联交易，同一level内是无关联交易
type transaction struct {
	ctx     context.Context
	cancel  context.CancelFunc

	total int
	split int
	concurrency int

	accounts    []*account.Account
	// address => accounts
	addresses   map[string]*account.Account

	level       int
	// 保存生成的中间结果：level+txs
	queue       chan levelTx
	// 产出结果：根据level隔离的交易队列
	levelQueue  []chan []*pb.Transaction
}

type levelTx struct {
	level   int
	txs     []*pb.Transaction
}

func NewTransaction(total, concurrency, split int, accounts []*account.Account, bootTx *pb.Transaction) (*transaction, error) {
	if total <= 1 || split <= 1 {
		return nil, fmt.Errorf("split must gt 1: total=%v, split=%v", total, split)
	}

	if len(accounts) < split {
		return nil, fmt.Errorf("accounts not enough: split=%d, accounts=%d", split, len(accounts))
	}

	addresses := make(map[string]*account.Account, len(accounts))
	for _, ak := range accounts {
		addresses[ak.Address] = ak
	}

	// 计算分裂层数
	level := math.Log10(float64(total)) / math.Log10(float64(split))

	ctx, cancel := context.WithCancel(context.Background())
	t := &transaction{
		ctx: ctx,
		cancel: cancel,

		total: total,
		split: split,
		concurrency: concurrency,
		level: int(math.Ceil(level)),

		accounts: accounts,
		addresses: addresses,
	}

	log.Printf("transaction: total=%d split=%d level=%d, concurrency=%d", t.total, t.split, t.level, t.concurrency)

	t.queue = make(chan levelTx, t.concurrency)
	t.levelQueue = make([]chan []*pb.Transaction, t.level+1)
	for i := 0; i <= t.level; i++ {
		t.levelQueue[i] = make(chan []*pb.Transaction, 10*split)
	}
	go func() {
		select {
		case <-t.ctx.Done():
			for i := 0; i <= t.level; i++ {
				close(t.levelQueue[i])
			}
		}
	}()

	go t.producer(bootTx)
	return t, nil
}

// 将所有level队列的交易合并到一个队列
func (t *transaction) Generate() chan []*pb.Transaction {
	queue := make(chan []*pb.Transaction, 10*t.concurrency)
	go func() {
		t.merge(0, queue)
		close(queue)
	}()
	return queue
}

func (t *transaction) merge(level int, queue chan []*pb.Transaction) {
	if level > t.level{
		return
	}

	txs, ok := <- t.levelQueue[level]
	if !ok {
		return
	}

	queue <- txs
	if level == 0 {
		time.Sleep(time.Second)
	}
	for i := 0; i < len(txs); i++ {
		t.merge(level+1, queue)
	}
}

// 自定义消费交易的方式
func (t *transaction) Consumer(consume func(level int, txs []*pb.Transaction) error) {
	wg := new(sync.WaitGroup)
	wg.Add(len(t.levelQueue))
	dealer := func(level int, ch chan []*pb.Transaction) {
		defer wg.Done()
		for txs := range ch {
			if err := consume(level, txs); err != nil {
				return
			}
		}
	}

	for i, ch := range t.levelQueue {
		go dealer(i, ch)
	}
	wg.Wait()
}

func (t *transaction) Level() int {
	return t.level
}

// 分裂交易：递归分裂交易，深度优先
func (t *transaction) splitRecursion(tx *pb.Transaction, level int) {
	if level >= t.level {
		return
	}

	select {
	case <-t.ctx.Done():
		return
	default:
	}

	childTxs := t.generate(tx, t.split)
	t.queue <- levelTx{
		level: level,
		txs: childTxs,
	}
	for _, childTx := range childTxs {
		t.splitRecursion(childTx, level+1)
	}
}

// 生产交易
func (t *transaction) producer(tx *pb.Transaction) {
	ctx, cancel := context.WithCancel(t.ctx)
	leafQueue := make(chan *pb.Transaction, t.concurrency)
	go t.splitRecursion(tx, 0)
	go t.leafTx(leafQueue, cancel)

	for {
		select {
		case <-ctx.Done():
			close(leafQueue)
			return
		case ltx := <-t.queue:
			//t.store(txs, level)
			t.levelQueue[ltx.level] <- ltx.txs
			if ltx.level + 1 == t.level {
				for _, tx := range ltx.txs {
					leafQueue <- tx
				}
			}
		}
	}
}

// 并发处理最后一层交易，提高性能
func (t *transaction) leafTx(leafQueue chan *pb.Transaction, cancel context.CancelFunc) {
	var count int64
	wg := new(sync.WaitGroup)
	wg.Add(t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		go func() {
			defer wg.Done()
			for tx := range leafQueue {
				leafTxs := t.generate(tx, 1)

				total := atomic.AddInt64(&count, int64(len(leafTxs)))
				if total%100000 == 0 {
					log.Printf("count=%d\n", total)
				}

				if total >= int64(t.total+t.split) {
					cancel()
					return
				}

				//t.store(txs, txsLevel.level+1)
				t.levelQueue[t.level] <- leafTxs
				if total >= int64(t.total) {
					cancel()
					return
				}
			}
		}()
	}
	wg.Wait()
	t.cancel()
}

// 生成子交易
func (t *transaction) generate(tx *pb.Transaction, split int) []*pb.Transaction {
	txs := make([]*pb.Transaction, len(tx.TxOutputs))
	for i, txOutput := range tx.TxOutputs {
		if txOutput == nil {
			continue
		}

		ak := t.addresses[string(txOutput.ToAddr)]
		if ak == nil {
			ak = BankAK
		}

		input := &pb.TxInput{
			RefTxid: tx.Txid,
			RefOffset: int32(i),
			FromAddr: txOutput.ToAddr,
			Amount: txOutput.Amount,
		}
		output := &pb.TxOutput{
			ToAddr: []byte(t.accounts[i].Address),
			Amount: txOutput.Amount,
		}
		newTx := TransferTx(input, output, split)
		txs[i] = SignTx(newTx, ak)
	}

	return txs
}

func BootTx(txid, address, amountStr string) (*pb.Transaction, error) {
	if len(txid) <= 0 {
		return nil, fmt.Errorf("txID not exist")
	}

	amount, ok := big.NewInt(0).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("boot tx amount error")
	}

	tx := &pb.Transaction{
		Txid: utils.DecodeId(txid),
		TxOutputs: []*pb.TxOutput{
			{
				ToAddr: []byte(address),
				Amount: amount.Bytes(),
			},
		},
	}

	return tx, nil
}

func TransferTx(input *pb.TxInput, output *pb.TxOutput, split int) *pb.Transaction {
	tx := &pb.Transaction{
		Version:   3,
		//Desc:      RandBytes(200),
		Nonce:     strconv.FormatInt(time.Now().UnixNano(), 36),
		Timestamp: time.Now().UnixNano(),
		Initiator: string(input.FromAddr),
		TxInputs: []*pb.TxInput{input},
		TxOutputs: Split(output, split),
	}

	return tx
}

func Split(txOutput *pb.TxOutput, split int) []*pb.TxOutput {
	if split <= 1 {
		return []*pb.TxOutput{txOutput}
	}

	total := big.NewInt(0).SetBytes(txOutput.Amount)
	if big.NewInt(int64(split)).Cmp(total) == 1 {
		// log.Printf("split utxo <= balance required")
		panic("amount not enough")
		return []*pb.TxOutput{txOutput}
	}

	amount := big.NewInt(0)
	amount.Div(total, big.NewInt(int64(split)))

	output := pb.TxOutput{}
	output.Amount = amount.Bytes()
	output.ToAddr = txOutput.ToAddr

	rest := total
	txOutputs := make([]*pb.TxOutput, 0, split+1)
	for i := 1; i < split && rest.Cmp(amount) == 1; i++ {
		tmpOutput := output
		txOutputs = append(txOutputs, &tmpOutput)
		rest.Sub(rest, amount)
	}
	output.Amount = rest.Bytes()
	txOutputs = append(txOutputs, &output)

	return txOutputs
}
