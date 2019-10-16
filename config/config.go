package config

import (
	"encoding/json"
	"os"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

type Config struct {
	// 压测的区块链类型
	BCType common.BlockChain `json:"type"`

	// 工作进程个数/并发数
	WorkerNum int `json:"workNum"`

	Host string `json:"host"`

	Nodes []string `json:"nodes"`

	Chain string `json:"chain"`

	Crypto string `json:"crypto"`

	// 压测模式，一共有两种模式
	// local模式：master和worker在一台机器上；
	// 分布式模式：master和worker散步在不同的机器上，依赖于中间件进行通信和管理
	Mode          common.TestMode `json:"mode"`
	Broker        string        `json:"broker"`
	ResultBackend string        `json:"resultBackend"`
	PubSubChan    string        `json:"pubSubChan"`

	Rounds []struct {
		Label    common.CaseType `json:"label"`
		Number   []int         `json:"number,omitempty"`
		Duration []int         `json:"duration,omitempty"`
	}
}

func GetBenchMsgFromConfigFile(fileName string) []*common.BenchMsg {
	conf := ParseConfig(fileName)
	return GetBenchMsgFromConf(conf)
}

func GetBenchMsgFromConf(conf *Config) []*common.BenchMsg {
	if conf == nil {
		log.ERROR.Printf("get %v conf, Config should not empty!", conf)
		return nil
	}

	benchMsg := make([]*common.BenchMsg, 0)
	for _, round := range conf.Rounds {
		if len(round.Duration) > 0 {
			for _, v := range round.Duration {
				msg := &common.BenchMsg{
					TestCase: common.TestCase{
						BCType: conf.BCType,
						Label: round.Label,
					},
					TxDuration: v,
					Parallel: conf.WorkerNum,
					Env: common.TestEnv{
						Host: conf.Host,
						Nodes: conf.Nodes,
						Crypto: conf.Crypto,
						Duration: v,
						Chain: conf.Chain,
					},
				}

				benchMsg = append(benchMsg, msg)
			}
		} else {
			for _, v := range round.Number {
				msg := &common.BenchMsg{
					TestCase: common.TestCase{
						BCType: conf.BCType,
						Label:  round.Label,
					},
					TxNumber: v,
					Parallel: conf.WorkerNum,
					Env: common.TestEnv{
						Host: conf.Host,
						Nodes: conf.Nodes,
						Crypto: conf.Crypto,
						Batch: v,
						Chain: conf.Chain,
					},
				}

				benchMsg = append(benchMsg, msg)
			}
		}
	}

	SetCallBack(benchMsg)

	return benchMsg
}

func ParseConfig(fileName string) *Config {
	file, err := os.Open(fileName)
	if err != nil {
		log.ERROR.Printf("encount err: %s", err)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	conf := Config{}

	if err := decoder.Decode(&conf); err != nil {
		log.ERROR.Printf("unmarshal config file <%s> error: %s", fileName, err)
		return nil
	}

	return &conf
}
