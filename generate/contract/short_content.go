package contracts

import (
	"fmt"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"strconv"
)

type shortContent struct {
	length int

	client *xuper.XClient
	config *ContractConfig
}

func NewShortContent(client *xuper.XClient, config *ContractConfig) (Contract, error) {
	t := &shortContent{
		client: client,
		config: config,
	}

	lengthStr, ok := config.Args["length"]
	if !ok {
		return nil, fmt.Errorf("params error: short content length not exist")
	}

	n, err := strconv.ParseUint(lengthStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("params error: %v, lenght=%s", err, lengthStr)
	}

	if n <= 0 || n > 3000 {
		t.length = 64
	} else {
		t.length = int(n)
	}

	return t, nil
}

func (t *shortContent) Deploy(from *account.Account, name string, code []byte, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	args = map[string]string{
		"creator": from.Address,
	}

	return t.client.DeployWasmContract(from, name, code, args, opts...)
}

// user_id: string, 用户名
// topic: string 类别(不超过36个字符)
// title: string, 标题(不超过100个字符)
// content: 具体内容(不超过3000个字符)
func (t *shortContent) Invoke(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	args = map[string]string{
		"user_id": `xuperos`,
		"topic": from.Address,
		"title": fmt.Sprintf("title_%d_%s", t.length, generate.RandBytes(16)),
		"content": string(generate.RandBytes(t.length)),
	}
	return t.client.InvokeWasmContract(from, name, method, args, opts...)
}

func (t *shortContent) Query(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	return t.client.QueryWasmContract(from, name, method, args, opts...)
}

func init() {
	RegisterContract("short_content", NewShortContent)
}
