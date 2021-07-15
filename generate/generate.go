package generate

import (
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

// 离线生成tx接口
type Generator interface {
	Generate() chan []*pb.Transaction
}

// sdk
type SDKGenerator interface {
	Init() error
	Generate() []chan *pb.Transaction
}

///////////////////////////////////////////////////////////////////////////////
// 离线生成交易: not SelectUTXO
type offlineTransfer struct {
	host        string
	concurrency int

	client      *xuper.XClient
	accounts    []*account.Account
	bootTxs     []*pb.Transaction
}

func NewOfflineTransfer(host string, concurrency int) (SDKGenerator, error) {
	t := &offlineTransfer{
		host: host,
		concurrency: concurrency,
	}

	var err error
	t.accounts, err = LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	t.client, err = xuper.New(t.host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}

	return t, nil
}

// 初始化业务
func (t *offlineTransfer) Init() error {
	txs, err := Transfer(t.client, BankAK, t.accounts, "100000000", 1)
	if err != nil {
		return fmt.Errorf("transfer to test accounts error: %v", err)
	}

	t.bootTxs = txs
	return nil
}

func (t *offlineTransfer) Generate() []chan *pb.Transaction {
	queues := make([]chan *pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan *pb.Transaction, 1)
	}

	provider := func(i int) {
		loop := make(chan *pb.Transaction, 1)
		loop <- t.bootTxs[i]
		for tx := range loop {
			ak := t.accounts[i]
			child := Fork(tx, ak)

			loop <- child
			queues[i] <- child
		}
	}

	for i := 0; i < t.concurrency; i++ {
		go provider(i)
	}

	return queues
}

func Fork(tx *pb.Transaction, ak *account.Account) *pb.Transaction {
	txOutput := tx.TxOutputs[0]
	input := &pb.TxInput{
		RefTxid: tx.Txid,
		RefOffset: 0,
		FromAddr: txOutput.ToAddr,
		Amount: txOutput.Amount,
	}
	output := &pb.TxOutput{
		ToAddr: []byte(ak.Address),
		Amount: txOutput.Amount,
	}
	newTx := TransferTx(input, output, 1)
	childTx := SignTx(newTx, ak)
	return childTx
}

///////////////////////////////////////////////////////////////////////////////
// 调用sdk生成tx
type transfer struct {
	host        string
	concurrency int

	client      *xuper.XClient
	accounts    []*account.Account
}

func NewTransfer(host string, concurrency int) (SDKGenerator, error) {
	t := &transfer{
		host: host,
		concurrency: concurrency,
	}

	var err error
	t.accounts, err = LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	t.client, err = xuper.New(t.host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}

	return t, nil
}

func (t *transfer) Init() error {
	_, err := Transfer(t.client, BankAK, t.accounts, "100000000", 10)
	if err != nil {
		return fmt.Errorf("transfer to test accounts error: %v", err)
	}

	return nil
}

func (t *transfer) Generate() []chan *pb.Transaction {
	queues := make([]chan *pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan *pb.Transaction, 1)
	}

	length := len(t.accounts)
	var counter int64
	provider := func(i int) {
		loop := make(chan *account.Account, 1)
		loop <- t.accounts[i]
		for from := range loop {
			to := t.accounts[rand.Intn(length)]
			tx, err := t.client.Transfer(from, to.Address, "10", xuper.WithNotPost())
			if err != nil {
				c := atomic.AddInt64(&counter, 1)
				log.Printf("generate tx error: %v, address=%s, counter=%d", err, from.Address, c)
				if c > 3 {
					return
				}

				time.Sleep(100*time.Millisecond)
				loop <- from
				continue
			}

			loop <- from
			queues[i] <- tx.Tx
		}
	}

	for i := 0; i < t.concurrency; i++ {
		go provider(i)
	}

	return queues
}
