package lib

import (
	"bytes"
	"fmt"
	"log"
	"math/big"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
)

// 转账给初始化账户
func InitTransfer(client *xuper.XClient, from *account.Account, accounts []*account.Account, amount string, split int) ([]*pb.Transaction, error) {
	txs := make([]*pb.Transaction, 0, len(accounts))
	for _, to := range accounts {
		tx, err := TransferWithSplit(client, from, to.Address, amount, split)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
		//log.Printf("address=%s, txid=%x", to.Address, tx.Txid)
	}

	log.Printf("transfer done")
	return txs, nil
}

// 分裂账户余额，避免并发请求时 "no enough money"
func SplitTx(host string, ak *account.Account, amount string, split int) error {
	client, err := xuper.New(host)
	if err != nil {
		return fmt.Errorf("new xuper client error: %v", err)
	}

	tx, err := TransferWithSplit(client, ak, ak.Address, amount, split)
	if err != nil {
		return err
	}

	log.Printf("split tx done, tx=%x", tx.Txid)
	return nil
}

// 转账
func TransferWithSplit(client *xuper.XClient, from *account.Account, to string, amount string, split int) (*pb.Transaction, error) {
	if amount == "" || split <= 0 {
		return nil, fmt.Errorf("params error: amount=%s, split=%d", amount, split)
	}

	tx, err := client.Transfer(from, to, amount, xuper.WithNotPost())
	if err != nil {
		return nil, err
	}

	txOutputs := make([]*pb.TxOutput, 0, len(tx.Tx.TxOutputs)+split)
	for _, txOutput := range tx.Tx.TxOutputs {
		if bytes.Equal(txOutput.ToAddr, []byte(to)) {
			txOutputs = append(txOutputs, SplitUTXO(txOutput, split)...)
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

	return tx.Tx, nil
}

func SplitUTXO(txOutput *pb.TxOutput, split int) []*pb.TxOutput {
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

