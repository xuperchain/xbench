package cases

import (
	"fmt"
	"github.com/xuperchain/xbench/lib"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"math/rand"
)

// 调用sdk生成tx
type transfer struct {
	host        string
	concurrency int
	split       int

	client      *xuper.XClient
	accounts    []*account.Account
}

func NewTransfer(config *Config) (Generator, error) {
	t := &transfer{
		host: config.Host,
		concurrency: config.Concurrency,
		split: 10,
	}

	var err error
	t.accounts, err = lib.LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	t.client, err = xuper.New(t.host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}

	log.Printf("generate: type=transfer, concurrency=%d", t.concurrency)
	return t, nil
}

func (t *transfer) Init() error {
	_, err := lib.Transfer(t.client, lib.BankAK, t.accounts, "100000000", t.split)
	if err != nil {
		return fmt.Errorf("transfer to test accounts error: %v", err)
	}

	return nil
}

func (t *transfer) Generate(id int) (*pb.Transaction, error) {
	from := t.accounts[id]
	to := t.accounts[rand.Intn(len(t.accounts))]
	tx, err := t.client.Transfer(from, to.Address, "10", xuper.WithNotPost())
	if err != nil {
		log.Printf("generate tx error: %v, address=%s", err, from.Address)
		return nil, err
	}

	return tx.Tx, nil
}

func init() {
	RegisterGenerator(CaseTransfer, NewTransfer)
}
