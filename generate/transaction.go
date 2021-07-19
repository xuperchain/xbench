package generate

import (
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"math/big"
	"strconv"
	"sync/atomic"
	"time"
)

// 离线生成交易: no SelectUTXO
type transaction struct {
	host        string
	total       int
	concurrency int

	client      *xuper.XClient
	accounts    []*account.Account
	bootTxs     []*pb.Transaction
}

func NewTransaction(config *Config) (Generator, error) {
	t := &transaction{
		host: config.Host,
		total: config.Total,
		concurrency: config.Concurrency,
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

	log.Printf("generate: type=transaction, total=%d, concurrency=%d", t.total, t.concurrency)
	return t, nil
}

func (t *transaction) Init() error {
	txs, err := Transfer(t.client, BankAK, t.accounts, "100000000", 1)
	if err != nil {
		return fmt.Errorf("transfer to test accounts error: %v", err)
	}

	t.bootTxs = txs
	return nil
}

func (t *transaction) Generate() []chan *pb.Transaction {
	queues := make([]chan *pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan *pb.Transaction, t.concurrency)
	}

	var count int64
	total := t.total / t.concurrency
	provider := func(i int) {
		ak := t.accounts[i]
		tx := t.bootTxs[i]
		for j := 0; j < total; j++ {
			child := Fork(tx, ak)
			queues[i] <- child
			tx = child
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
