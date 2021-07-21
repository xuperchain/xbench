package cases

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/xuperchain/xbench/lib"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuperchain/service/pb"
)

type evidence struct {
	host        string
	concurrency int
	length      int

	accounts    []*account.Account
}

func NewEvidence(config *Config) (Generator, error) {
	t := &evidence{
		host: config.Host,
		concurrency: config.Concurrency,
	}

	var err error
	t.length, err = strconv.Atoi(config.Args["length"])
	if err != nil {
		return nil, fmt.Errorf("evidence length error: %v", err)
	}

	t.accounts, err = lib.LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	log.Printf("generate: type=evidence, concurrency=%d, length=%d", t.concurrency, t.length)
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

	lib.SignTx(tx, ak)
	return tx
}

func init() {
	RegisterGenerator(CaseEvidence, NewEvidence)
}
