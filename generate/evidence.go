package generate

import (
	"fmt"
	"github.com/xuperchain/xbench/lib"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"strconv"
	"time"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
)

type evidence struct {
	host        string
	total       int
	concurrency int
	length      int
	batch       int

	accounts    []*account.Account
}

func NewEvidence(config *Config) (Generator, error) {
	t := &evidence{
		host: config.Host,
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

func (t *evidence) Generate(id int) (*pb.Transaction, error) {
	ak := t.accounts[id]
	tx := EvidenceTx(ak, t.length)
	return tx, nil
}

func EvidenceTx(ak *account.Account, length int) *pb.Transaction {
	tx := &pb.Transaction{
		Version:   3,
		Desc:      lib.RandBytes(length),
		Nonce:     strconv.FormatInt(time.Now().UnixNano(), 36),
		Timestamp: time.Now().UnixNano(),
		Initiator: ak.Address,
	}

	SignTx(tx, ak)
	return tx
}

func init() {
	RegisterGenerator(BenchmarkEvidence, NewEvidence)
}
