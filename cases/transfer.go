package cases

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xbench/lib"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
)

// 调用sdk生成tx
type transfer struct {
	host        string
	concurrency int
	split       int
	amount      string

	client      *xuper.XClient
	accounts    []*account.Account
}

func NewTransfer(config *Config) (Generator, error) {
	t := &transfer{
		host: config.Host,
		concurrency: config.Concurrency,
		split: 10,
		amount: config.Args["amount"],
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
	_, err := lib.InitTransfer(t.client, lib.Bank, t.accounts, t.amount, t.split)
	if err != nil {
		return fmt.Errorf("transfer to test accounts error: %v", err)
	}

	return nil
}

func (t *transfer) Generate(id int) (proto.Message, error) {
	from := t.accounts[id]
	to := t.accounts[rand.Intn(len(t.accounts))]
	tx, err := t.client.Transfer(from, to.Address, "100", xuper.WithNotPost())
	if err != nil {
		log.Printf("generate tx error: %v, address=%s", err, from.Address)
		return nil, err
	}

	return tx.Tx, nil
}

func init() {
	RegisterGenerator(CaseTransfer, NewTransfer)
}
