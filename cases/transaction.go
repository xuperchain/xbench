package cases

import (
	"fmt"
	"github.com/xuperchain/xbench/lib"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"strconv"
	"time"
)

// 离线生成交易: no SelectUTXO
type transaction struct {
	host        string
	concurrency int

	client      *xuper.XClient
	accounts    []*account.Account
	bootTxs     []*pb.Transaction
}

func NewTransaction(config *Config) (Generator, error) {
	t := &transaction{
		host: config.Host,
		concurrency: config.Concurrency,
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

	log.Printf("generate: type=transaction, concurrency=%d", t.concurrency)
	return t, nil
}

func (t *transaction) Init() error {
	txs, err := lib.Transfer(t.client, lib.BankAK, t.accounts, "100000000", 1)
	if err != nil {
		return fmt.Errorf("transfer to test accounts error: %v", err)
	}

	t.bootTxs = txs
	return nil
}

func (t *transaction) Generate(id int) (*pb.Transaction, error){
	tx := t.bootTxs[id]
	ak := t.accounts[id]
	child := Fork(tx, ak)
	t.bootTxs[id] = child
	return child, nil
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
	newTx := TransactionTx(input, output, 1)
	childTx := lib.SignTx(newTx, ak)
	return childTx
}

func TransactionTx(input *pb.TxInput, output *pb.TxOutput, split int) *pb.Transaction {
	tx := &pb.Transaction{
		Version:   3,
		//Desc:      RandBytes(200),
		Nonce:     strconv.FormatInt(time.Now().UnixNano(), 36),
		Timestamp: time.Now().UnixNano(),
		Initiator: string(input.FromAddr),
		TxInputs: []*pb.TxInput{input},
		TxOutputs: lib.SplitUTXO(output, split),
	}

	return tx
}

func init() {
	RegisterGenerator(CaseTransaction, NewTransaction)
}
