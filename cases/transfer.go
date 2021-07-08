package cases

import (
	"github.com/xuperchain/xuper-sdk-go/account"
	"github.com/xuperchain/xuperchain/service/pb"
	"math/big"
	"strconv"
)

func InvokeContract(request *pb.InvokeRequest, from *account.Account) (*pb.Transaction, error) {
	out, err := PreExecWithSelectUTXOContract(request, from)
	if err != nil {
		return nil, err
	}
	tx := GenerateTx(out, from)
	return SignTx(tx, from, ""), nil
}

func Transfer(from *account.Account, to string, amount string) (*pb.Transaction, error) {
	need, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return nil, err
	}
	out, err := PreExecWithSelectUTXO(from, need)
	if err != nil {
		return nil, err
	}

	amt, _ := big.NewInt(0).SetString(amount, 10)
	out.Response.UtxoOutputs = append(out.Response.UtxoOutputs, &pb.TxOutput{
		ToAddr: []byte(to),
		Amount: amt.Bytes(),
	})
	tx := GenerateTx(out, from)
	return SignTx(tx, from, ""), nil
}

func TransferWithSplit(from *account.Account, to string, amount string, split int64) (*pb.Transaction, error) {
	need, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return nil, err
	}
	out, err := PreExecWithSelectUTXO(from, need)
	if err != nil {
		return nil, err
	}

	amt, _ := big.NewInt(0).SetString(amount, 10)
	out.Response.UtxoOutputs = append(out.Response.UtxoOutputs, &pb.TxOutput{
		ToAddr: []byte(to),
		Amount: amt.Bytes(),
	})

	tx := GenerateTx(out, from)
	if split > 1 {
		txOutputs := make([]*pb.TxOutput, 0, int(split))
		for _, output := range tx.TxOutputs {
			address := string(output.ToAddr)
			outputAmount := big.NewInt(0).SetBytes(output.Amount)
			total, _ := big.NewInt(0).SetString(amount, 10)
			if address == to && outputAmount.Cmp(total) == 0 {
				outputs, err := SplitOutputs(to, total, split)
				if err != nil {
					return nil, err
				}

				txOutputs = append(txOutputs, outputs...)
			} else {
				txOutputs = append(txOutputs, output)
			}
		}
		tx.TxOutputs = txOutputs
	}
	return SignTx(tx, from, ""), nil
}