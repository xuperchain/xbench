package cases

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
)

type RandCase struct {
	common.TestCase
}

var (
	cnt map[int]int
	tms map[int]time.Duration
	bid string
	tid string
)

func (r RandCase) Init(args ...interface{}) error {
	parallel := args[0].(int)
	env := args[1].(common.TestEnv)
	lib.SetCrypto(env.Crypto)
	amount := env.Batch
	rand.Seed(time.Now().UnixNano())
	Bank = lib.InitBankAcct("")
	lib.BatchRetrieve(LAccts, parallel, env.Chain, env.Host)
	for i:=0; i<parallel; i++ {
		Accts[i] = lib.RetriveAcct(LAccts[i])
		if len(Clis) < parallel {
			cli := lib.Conn(env.Host, env.Chain)
			Clis = append(Clis, cli)
		}
		lib.Transplit(Bank, Accts[i].Address, amount, Clis[i])
	}
	rsp, _ := Clis[0].GetSystemStatus()
	for _, b := range(rsp.SystemsStatus.BcsStatus) {
		if b.Bcname == env.Chain {
			bid = hex.EncodeToString(b.UtxoMeta.LatestBlockid)
			binfo, _ := Clis[0].GetBlock(bid)
			tid = hex.EncodeToString(binfo.Block.Transactions[0].Txid)
		}
		break
	}
	cnt = make(map[int]int, 6)
	tms = make(map[int]time.Duration, 6)
	return nil
}

func (r RandCase) Run(seq int, args ...interface{}) error {
	s := time.Now()
	switch i := rand.Intn(6); i {
	case 0:
		lib.Trans(Accts[seq], Bank.Address, "1", Clis[seq])
		cnt[0]++
		tms[0] += time.Since(s)
	case 1:
		lib.Transplit(Accts[seq], Bank.Address, 1, Clis[seq])
		cnt[1]++
		tms[1] += time.Since(s)
	case 2:
		k := fmt.Sprintf("key_%d", seq)
		lib.InvokeContract(Bank, contract, "increase", k, Clis[seq])
		cnt[2]++
		tms[1] += time.Since(s)
	case 3:
		Clis[seq].GetBlock(bid)
		cnt[3]++
		tms[3] += time.Since(s)
	case 4:
		Clis[seq].QueryTx(tid)
		cnt[4]++
		tms[4] += time.Since(s)
	case 5:
		Clis[seq].GetBalance(Accts[0].Address)
		cnt[5]++
		tms[5] += time.Since(s)
	}
	return nil
}

func (r RandCase) End(args ...interface{}) error {
	fmt.Println(cnt)
	fmt.Println(tms)
	return nil
}
