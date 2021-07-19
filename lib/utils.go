package lib

import (
	"encoding/json"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/lib/crypto/client"
	"io"
	"math/rand"
	"strconv"
	"strings"
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

func FormatTx(tx *pb.Transaction) []byte {
	t := FromPBTx(tx)
	data, _ := json.MarshalIndent(t, "", "  ")
	return data
}

func WorkID(workID string) int {
	workIdStr := strings.Split(workID[1:], "c")[0]
	workId, _ := strconv.Atoi(workIdStr)
	return workId
}

// 从chan中获取tx写文件
func WriteFile(queue chan *pb.Transaction, w io.Writer, total int) error {
	count := 0
	e := json.NewEncoder(w)
	for tx := range queue {
		err := e.Encode(tx)
		if err != nil {
			return err
		}

		count++
		if total > 0 && count >= total {
			return nil
		}
	}

	return nil
}

func ReadFile(r io.Reader, queue chan *pb.Transaction, total int) error {
	count := 0
	dec := json.NewDecoder(r)
	for {
		var m pb.Transaction
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		queue <- &m

		count++
		if total > 0 && count >= total {
			return nil
		}
	}

	return nil
}
