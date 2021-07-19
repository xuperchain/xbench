package generate

import (
	"encoding/json"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/lib/crypto/client"
)

func SignTx(tx *pb.Transaction, from *account.Account) *pb.Transaction {
	crypto, _ := client.CreateCryptoClient("default")
	signTx, _ := common.ComputeTxSign(crypto, tx, []byte(from.PrivateKey))
	signInfo := &pb.SignatureInfo{
		PublicKey: from.PublicKey,
		Sign: signTx,
	}
	tx.InitiatorSigns = append(tx.InitiatorSigns, signInfo)
	tx.Txid, _ = common.MakeTxId(tx)
	return tx
}

func FormatTx(tx *pb.Transaction) []byte {
	t := FromPBTx(tx)
	data, _ := json.MarshalIndent(t, "", "  ")
	return data
}
