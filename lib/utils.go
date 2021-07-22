package lib

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/lib/crypto/client"
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// 产生随机字符串
func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return b
}

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

func WorkID(workID string) int {
	workIdStr := strings.Split(workID[1:], "c")[0]
	workId, _ := strconv.Atoi(workIdStr)
	return workId
}
