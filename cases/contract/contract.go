package contracts

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuperchain/service/pb"
)

type Contract interface {
	Init() error
	GenerateTx(run *runner.CallData) (*pb.Transaction, error)
}

type Config struct {
	Host            string

	// 合约地址
	ContractAccount string
	// 合约路径
	CodePath        string

	// 模块名：wasm/native/evm
	ModuleName      string
	// 合约名
	ContractName    string
	// 方法名
	MethodName      string

	// 自定义参数
	Args            map[string]string
}

type NewContract func (config *Config, accounts []*account.Account) (Contract, error)

// 注册合约
var contracts = make(map[string]NewContract, 8)

func RegisterContract(name string, contract NewContract) {
	contracts[name] = contract
}

func GetContract(name string) (NewContract, error) {
	if newContract, ok := contracts[name]; ok {
		return newContract, nil
	}

	return nil, fmt.Errorf("contract not exist, contract_name=%s", name)
}
