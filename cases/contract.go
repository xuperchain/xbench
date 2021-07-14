package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	contracts "github.com/xuperchain/xbench/cases/contract"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuperchain/service/pb"
)

const BlockChain = "xuper"

type Contract struct {
	concurrency     int

	config   *contracts.Config
	contract contracts.Contract
	accounts []*account.Account
}

func NewContract(config runner.Config) (Provider, error) {
	t := &Contract{
		concurrency: int(config.C),

		config: &contracts.Config{
			ModuleName: config.Tags["module_name"],
			ContractName: config.Tags["contract_name"],
			ContractAccount: config.Tags["contract_account"],
			MethodName: config.Tags["method_name"],
			Host: config.Host,
			Args: config.Tags,
		},
	}

	var err error
	t.accounts, err = generate.LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	newContract, err := contracts.GetContract(t.config.ContractName)
	if err != nil {
		return nil, fmt.Errorf("get newContract error: %v", err)
	}

	t.contract, err = newContract(t.config, t.accounts)
	if err != nil {
		return nil, fmt.Errorf("new contract error: %v", err)
	}

	return t, nil
}

func (t *Contract) DataProvider(run *runner.CallData) ([]*dynamic.Message, error) {
	tx, err := t.contract.GenerateTx(run)
	if err != nil {
		return nil, fmt.Errorf("contract generate tx error: %v", err)
	}

	msg := &pb.TxStatus{
		Bcname: BlockChain,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Txid:   tx.Txid,
	}
	dynamicMsg, err := dynamic.AsDynamicMessage(msg)
	if err != nil {
		return nil, err
	}

	return []*dynamic.Message{dynamicMsg}, nil
}

func init() {
	RegisterProvider("contract", NewContract)
}
