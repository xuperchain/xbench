package lib

import (
	"bytes"
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"log"
	"math/big"
)

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

		txs = append(txs, tx.Tx)
		//log.Printf("address=%s, txid=%x", to.Address, tx.Tx.Txid)
	}

	log.Printf("transfer done")
	return txs, nil
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
	txOutputs = append(txOutputs, SplitUTXO(tx.Tx.TxOutputs[0], split)...)
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

