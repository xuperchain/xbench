package contracts

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
)

type counter struct {
	config      *Config
	accounts    []*account.Account
	client      *xuper.XClient
}

func NewCounter(config *Config, accounts []*account.Account) (Contract, error) {
	t := &counter{
		config: config,
		accounts: accounts,
	}

	client, err := xuper.New(t.config.Host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}

	t.client = client
	return t, nil
}

func (t *counter) Init() error {
	return nil
}

func (t *counter) GenerateTx(run *runner.CallData) (*pb.Transaction, error) {
	from := t.accounts[run.RequestNumber%int64(len(t.accounts))]
	from.SetContractAccount(t.config.ContractAccount)

	args := map[string]string{
		"key": fmt.Sprintf("test_%s", run.WorkerID),
	}

	tx, err := t.client.InvokeWasmContract(from, t.config.ContractName, t.config.MethodName, args, xuper.WithNotPost())
	if err != nil {
		return nil, err
	}

	return tx.Tx, nil
}

func init() {
	RegisterContract("counter", NewCounter)
}
