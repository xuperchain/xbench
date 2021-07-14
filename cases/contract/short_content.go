package contracts

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"strconv"
)

type shortContent struct{
	config      *Config
	accounts    []*account.Account
	client      *xuper.XClient

	length      int
}

func NewShortContent(config *Config, accounts []*account.Account) (Contract, error) {
	t := &shortContent{
		config: config,
		accounts: accounts,
	}

	if lengthStr, ok := config.Args["length"]; ok {
		n, _ := strconv.ParseUint(lengthStr, 10, 64)
		if n <= 0 || n > 3000 {
			t.length = 64
		} else {
			t.length = int(n)
		}
	}

	client, err := xuper.New(t.config.Host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}
	t.client = client
	return t, nil
}

func (t *shortContent) Init() error {
	return nil
}

//
// user_id: string, 用户名
// topic: string 类别(不超过36个字符)
// title: string, 标题(不超过100个字符)
// content: 具体内容(不超过3000个字符)
//
func (t *shortContent) GenerateTx(run *runner.CallData) (*pb.Transaction, error) {
	from := t.accounts[run.RequestNumber%int64(len(t.accounts))]
	from.SetContractAccount(t.config.ContractAccount)

	args := map[string]string{
		"user_id": `xuperos`,
		"topic": run.WorkerID,
		"title": fmt.Sprintf("title_%d_%s", t.length, generate.RandBytes(16)),
		"content": string(generate.RandBytes(t.length)),
	}

	tx, err := t.client.InvokeWasmContract(from, t.config.ContractName, t.config.MethodName, args, xuper.WithNotPost())
	if err != nil {
		return nil, err
	}

	return tx.Tx, nil
}

func init() {
	RegisterContract("short_content", NewShortContent)
}
