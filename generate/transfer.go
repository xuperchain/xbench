package generate

import (
	"bytes"
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"math/rand"
	"sync/atomic"
)

// 调用sdk生成tx
type transfer struct {
	host        string
	total       int
	concurrency int
	split       int

	client      *xuper.XClient
	accounts    []*account.Account
}

func NewTransfer(config *Config) (Generator, error) {
	t := &transfer{
		host: config.Host,
		total: int(float32(config.Total)*1.1),
		concurrency: config.Concurrency,
		split: 10,
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

	log.Printf("generate: type=transfer, total=%d, concurrency=%d", t.total, t.concurrency)
	return t, nil
}

func (t *transfer) Init() error {
	_, err := Transfer(t.client, BankAK, t.accounts, "100000000", t.split)
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

	var count int64
	length := len(t.accounts)
	total := t.total / t.concurrency
	provider := func(i int) {
		from := t.accounts[i]
		for j := 0; j < total; j++ {
			to := t.accounts[rand.Intn(length)]
			tx, err := t.client.Transfer(from, to.Address, "10", xuper.WithNotPost())
			if err != nil {
				log.Printf("generate tx error: %v, address=%s", err, from.Address)
				return
			}

			queues[i] <- tx.Tx
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

// 分裂账户余额，避免并发请求时 "no enough money"
func SplitTx(host string, ak *account.Account, split int) error {
	client, err := xuper.New(host)
	if err != nil {
		return fmt.Errorf("new xuper client error: %v", err)
	}

	amount := fmt.Sprintf("%d00000000", split)
	tx, err := client.Transfer(ak, ak.Address, amount, xuper.WithNotPost())
	if err != nil {
		return fmt.Errorf("transfer tx error: %v", err)
	}
	
	txOutputs := make([]*pb.TxOutput, 0, len(tx.Tx.TxOutputs)+split)
	txOutputs = append(txOutputs, Split(tx.Tx.TxOutputs[0], split)...)
	txOutputs = append(txOutputs, tx.Tx.TxOutputs[1:]...)

	tx.DigestHash = nil
	tx.Tx.TxOutputs = txOutputs
	tx.Tx.AuthRequireSigns = nil
	tx.Tx.InitiatorSigns = nil
	err = tx.Sign(ak)
	if err != nil {
		return fmt.Errorf("sign error: %v", err)
	}

	tx, err = client.PostTx(tx)
	if err != nil {
		return fmt.Errorf("post tx error: %v", err)
	}

	log.Printf("split tx done, tx=%x", tx.Tx.Txid)
	return nil
}

// 转账给初始化账户
func Transfer(client *xuper.XClient, from *account.Account, accounts []*account.Account, amount string, split int) ([]*pb.Transaction, error) {
	log.Printf("transfer start")

	txs := make([]*pb.Transaction, 0, len(accounts))
	for _, to := range accounts {
		tx, err := client.Transfer(from, to.Address, amount, xuper.WithNotPost())
		if err != nil {
			return nil, err
		}

		txOutputs := make([]*pb.TxOutput, 0, len(tx.Tx.TxOutputs)+split)
		for _, txOutput := range tx.Tx.TxOutputs {
			if bytes.Equal(txOutput.ToAddr, []byte(to.Address)) {
				txOutputs = append(txOutputs, Split(txOutput, split)...)
			} else {
				txOutputs = append(txOutputs, txOutput)
			}
		}

		tx.DigestHash = nil
		tx.Tx.TxOutputs = txOutputs
		tx.Tx.AuthRequireSigns = nil
		tx.Tx.InitiatorSigns = nil
		err = tx.Sign(from)
		if err != nil {
			return nil, err
		}

		tx, err = client.PostTx(tx)
		if err != nil {
			return nil, err
		}

		txs = append(txs, tx.Tx)
		//log.Printf("address=%s, txid=%x", to.Address, tx.Tx.Txid)
	}

	log.Printf("transfer done")
	return txs, nil
}
