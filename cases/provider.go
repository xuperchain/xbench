package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
)

type Provider interface {
	DataProvider(*runner.CallData) ([]*dynamic.Message, error)
}

type NewProvider func (config runner.Config) (Provider, error)

// 注册数据生成器
var providers = make(map[string]NewProvider, 8)
func RegisterProvider(name string, provider NewProvider) {
	providers[name] = provider
}

func MakeDataProvider(config runner.Config) (runner.DataProviderFunc, error) {
	if config.Tags == nil {
		return nil, fmt.Errorf("param nil")
	}

	benchType, ok := config.Tags["bench_type"]
	if !ok {
		return nil, fmt.Errorf("bench_type not exist")
	}

	newProvider, ok := providers[benchType]
	if !ok {
		return nil, fmt.Errorf("bench_type not exist: %s", benchType)
	}

	provider, err := newProvider(config)
	if err != nil {
		return nil, fmt.Errorf("new provider error: %v", err)
	}

	return provider.DataProvider, nil
}
