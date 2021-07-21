package cases

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	contracts "github.com/xuperchain/xbench/cases/contract"
	"github.com/xuperchain/xbench/lib"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
)

const WaitDeploy = 5 // 等待所有节点完成合约部署 5s

// 调用sdk生成tx
type contract struct {
	host        string
	concurrency int
	split       int
	amount      string
	waitDeploy  int

	config      *contracts.ContractConfig
	contract    contracts.Contract

	client      *xuper.XClient
	accounts    []*account.Account
}

func NewContract(config *Config) (Generator, error) {
	waitDeploy, _ := strconv.Atoi(config.Args["wait_deploy"])
	if waitDeploy <= 0 {
		waitDeploy = WaitDeploy
	}

	t := &contract{
		host: config.Host,
		concurrency: config.Concurrency,
		split: 10,
		amount: config.Args["amount"],
		waitDeploy: waitDeploy,

		config: &contracts.ContractConfig{
			ContractAccount: config.Args["contract_account"],
			CodePath: config.Args["code_path"],

			ModuleName: config.Args["module_name"],
			ContractName: config.Args["contract_name"],
			MethodName: config.Args["method_name"],
			Args: config.Args,
		},
	}

	var err error
	t.accounts, err = lib.LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	t.client, err = xuper.New(t.host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}

	t.contract, err = contracts.GetContract(t.config, t.client)
	if err != nil {
		return nil, fmt.Errorf("get contract error: %v, contract=%s", err, t.config.ContractName)
	}

	log.Printf("generate: type=contract, contract=%s, concurrency=%d", t.config.ContractName, t.concurrency)
	return t, nil
}

// 业务初始化
func (t *contract) Init() error {
	contractAccount := t.config.ContractAccount
	// 创建合约账户
	_, err := t.client.CreateContractAccount(lib.Bank, contractAccount)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("create account error: %v, account=%s", err, t.config.ContractAccount)
		}
		log.Printf("account already exists, account=%s", t.config.ContractAccount)
	}

	// 转账给合约账户
	_, err = t.client.Transfer(lib.Bank, contractAccount, t.amount)
	if err != nil {
		return fmt.Errorf("transfer to contract account error: %v, contractAccount=%s", err, contractAccount)
	}

	// 部署合约
	bank := lib.Bank
	if err := bank.SetContractAccount(contractAccount); err != nil {
		return err
	}
	code, err := ioutil.ReadFile(t.config.CodePath)
	if err != nil {
		return fmt.Errorf("read contract code error: %v", err)
	}
	_, err = t.contract.Deploy(bank, t.config.ContractName, code, t.config.Args)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("deploy contract error: %v, contract=%s", err, t.config.ContractName)
		}
		log.Printf("contract already exists, contract=%s", t.config.ContractName)
	}
	bank.RemoveContractAccount()
	log.Printf("deploy contract done")

	// 等待部署合约完成
	time.Sleep(time.Duration(t.waitDeploy)*time.Second)

	// 转账给调用合约的账户
	_, err = lib.InitTransfer(t.client, lib.Bank, t.accounts, t.amount, t.split)
	if err != nil {
		return fmt.Errorf("contract to test accounts error: %v", err)
	}

	log.Printf("init done")
	return nil
}

func (t *contract) Generate(id int) (*pb.Transaction, error) {
	from := t.accounts[id]
	args := map[string]string {
		"id": strconv.Itoa(id),
	}
	tx, err := t.contract.Invoke(from, t.config.ContractName, t.config.MethodName, args, xuper.WithNotPost())
	if err != nil {
		log.Printf("generate tx error: %v, address=%s", err, from.Address)
		return nil, err
	}

	return tx.Tx, nil
}

func init() {
	RegisterGenerator(CaseContract, NewContract)
}
