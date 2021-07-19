package provider

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
)

const BlockChain = "xuper"

const (
	CallPostTx = "pb.Xchain.PostTx"
	CallPreExec = "pb.Xchain.PreExec"
)

// provider：为 ghz 发压工具提供压测数据，不同的grpc接口有不同的实现
type Provider interface {
	DataProvider(*runner.CallData) ([]*dynamic.Message, error)
}

type NewProvider func (config *runner.Config) (Provider, error)

// 注册数据生成器
var providers = make(map[string]NewProvider, 8)
func RegisterProvider(name string, provider NewProvider) {
	providers[name] = provider
}

func GetProvider(name string, config *runner.Config) (Provider, error) {
	newProvider, ok := providers[name]
	if ok {
		return newProvider(config)
	}

	return nil, fmt.Errorf("call not exist: %s", name)
}

func NewDataProviderFunc(config *runner.Config) (runner.DataProviderFunc, error) {
	provider, err := GetProvider(config.Call, config)
	if err != nil {
		return nil, fmt.Errorf("get provider error: %v", err)
	}

	return provider.DataProvider, nil
}