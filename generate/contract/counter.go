package contracts

import (
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"strconv"
)

type counter struct {
	client *xuper.XClient
	config *ContractConfig
}

func NewCounter(client *xuper.XClient, config *ContractConfig) (Contract, error) {
	t := &counter{
		client: client,
		config: config,
	}

	return t, nil
}


func (t *counter) Deploy(from *account.Account, name string, code []byte, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	args = map[string]string{
		"creator": from.Address,
	}

	return t.client.DeployWasmContract(from, name, code, args, opts...)
}

func (t *counter) Invoke(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	requestId, _ := strconv.Atoi(args["requestId"])
	args = map[string]string{
		"key": fmt.Sprintf("test_%s_%d", from.Address, requestId%10),
	}
	return t.client.InvokeWasmContract(from, name, method, args, opts...)
}

func (t *counter) Query(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	return t.client.QueryWasmContract(from, name, method, args, opts...)
}

func init() {
	RegisterContract("counter", NewCounter)
}
