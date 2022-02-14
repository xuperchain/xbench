package contracts

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/sony/sonyflake"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
)

type contract struct {
	client *xuper.XClient
	config *ContractConfig
}

func NewContracts(config *ContractConfig, client *xuper.XClient) (Contract, error) {
	t := &contract{
		client: client,
		config: config,
	}

	return t, nil
}

func (t *contract) Deploy(from *account.Account, name string, code []byte, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	if len(t.config.Args["deploy_consts"]) != 0 {
		arrayArgs := strings.Split(t.config.Args["deploy_consts"], ",")
		for _, value := range arrayArgs {
			args[value] = t.config.Args[value]
		}
	}

	return t.client.DeployWasmContract(from, name, code, args, opts...)
}

func (t *contract) Invoke(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	// 变量数据组装
	if len(t.config.Args["invoke_vars"]) != 0 {
		rand.Seed(time.Now().UnixNano())
		settings := sonyflake.Settings{
			StartTime: time.Now(),
			MachineID: func() (u uint16, e error) {
				return uint16(rand.Int()), nil
			},
		}
		randId := sonyflake.NewSonyflake(settings)

		arrayArgs := strings.Split(t.config.Args["invoke_vars"], ",")
		for _, value := range arrayArgs {
			id, _ := randId.NextID()
			args[value] = strconv.FormatUint(id, 10)
		}
	}

	// 常量数据组装
	if len(t.config.Args["invoke_consts"]) != 0 {
		arrayArgs := strings.Split(t.config.Args["invoke_consts"], ",")
		for _, value := range arrayArgs {
			args[value] = t.config.Args[value]
		}
	}

	return t.client.InvokeWasmContract(from, name, method, args, opts...)
}

func (t *contract) Query(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error) {
	return nil, nil
}

func init() {
	RegisterContract("contracts", NewContracts)
}
