package contracts

import (
	"fmt"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
)

type Contract interface {
	Deploy(from *account.Account, name string, code []byte, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error)
	Invoke(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error)
	Query(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error)
}

type ContractConfig struct {
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

type NewContractFunc func (client *xuper.XClient, config *ContractConfig) (Contract, error)

// 注册合约
var contracts = make(map[string]NewContractFunc, 8)

func RegisterContract(name string, contract NewContractFunc) {
	contracts[name] = contract
}

func GetContract(name string) (NewContractFunc, error) {
	if newContract, ok := contracts[name]; ok {
		return newContract, nil
	}

	return nil, fmt.Errorf("contract not exist")
}

// 调用sdk生成tx
type contract struct {
	host        string
	total       int
	concurrency int
	split       int

	config      *ContractConfig
	contract    Contract

	client      *xuper.XClient
	accounts    []*account.Account
}

func NewContract(config *generate.Config) (generate.Generator, error) {
	t := &contract{
		host: config.Host,
		total: int(float32(config.Total)*1.1),
		concurrency: config.Concurrency,
		split: 10,

		config: &ContractConfig{
			ContractAccount: config.Args["contract_account"],
			CodePath: config.Args["code_path"],

			ModuleName: config.Args["module_name"],
			ContractName: config.Args["contract_name"],
			MethodName: config.Args["method_name"],
			Args: config.Args,
		},
	}

	var err error
	t.accounts, err = generate.LoadAccount(t.concurrency)
	if err != nil {
		return nil, fmt.Errorf("load account error: %v", err)
	}

	t.client, err = xuper.New(t.host)
	if err != nil {
		return nil, fmt.Errorf("new xuper client error: %v", err)
	}

	newContract, err := GetContract(t.config.ContractName)
	if err != nil {
		return nil, fmt.Errorf("get NewContractFunc error: %v, contract=%s", err, t.config.ContractName)
	}

	t.contract, err = newContract(t.client, t.config)
	if err != nil {
		return nil, fmt.Errorf("new contract error: %v, contract=%s", err, t.config.ContractName)
	}

	log.Printf("generate: type=contract, total=%d, concurrency=%d", t.total, t.concurrency)
	return t, nil
}

// 业务初始化
func (t *contract) Init() error {
	contractAccount := t.config.ContractAccount
	// 创建合约账户
	_, err := t.client.CreateContractAccount(generate.BankAK, contractAccount)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("create account error: %v, account=%s", err, t.config.ContractAccount)
		}
		log.Printf("account already exists, account=%s", t.config.ContractAccount)
	}

	// 转账给合约账户
	_, err = t.client.Transfer(generate.BankAK, contractAccount, "100000000")
	if err != nil {
		return fmt.Errorf("transfer to contract account error: %v, contractAccount=%s", err, contractAccount)
	}

	// 部署合约
	bank := generate.BankAK
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

	// 转账给调用合约的账户
	_, err = generate.Transfer(t.client, generate.BankAK, t.accounts, "100000000", t.split)
	if err != nil {
		return fmt.Errorf("contract to test accounts error: %v", err)
	}

	log.Printf("init done")
	return nil
}

func (t *contract) Generate() []chan *pb.Transaction {
	queues := make([]chan *pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan *pb.Transaction, 1)
	}

	var count int64
	total := t.total / t.concurrency
	provider := func(i int) {
		from := t.accounts[i]
		for j := 0; j < total; j++ {
			args := map[string]string {
				"concurrency": strconv.Itoa(i),
				"requestId": strconv.Itoa(j),
			}
			tx, err := t.contract.Invoke(from, t.config.ContractName, t.config.MethodName, args, xuper.WithNotPost())
			if err != nil {
				log.Printf("generate tx error: %v, address=%s", err, from.Address)
				return
			}

			queues[i] <- tx.Tx
			if (j+1) % t.concurrency == 0 {
				total := atomic.AddInt64(&count, int64(t.concurrency))
				if total%100000 == 0 {
					log.Printf("count=%d\n", total)
				}
			}
		}

		close(queues[i])
	}

	for i := 0; i < t.concurrency; i++ {
		go provider(i)
	}

	return queues
}
