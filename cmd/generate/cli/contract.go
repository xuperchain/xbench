package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/xuperchain/xbench/cases"
)

// 定义命令行参数变量
var envCfgPath string

// BenchCommand
type ContractCommand struct {
	cli    *Cli
	cmd    *cobra.Command
	config *Config
}

type Config struct {
	Total       int               `yaml:"total"`       // 交易总量
	Process     int               `yaml:"process"`     // 进程数
	Concurrency int               `yaml:"concurrency"` // 并发数
	Child       int               `yaml:"child"`       // 进程编号
	Output      string            `yaml:"output"`      // 产出路径
	Host        string            `yaml:"host"`
	Tags        map[string]string `yaml:"tags"`
}

func NewContractCommand(cli *Cli) *cobra.Command {
	t := new(ContractCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "contract",
		Short: "contract",
		RunE: func(cmd *cobra.Command, args []string) error {
			return t.ContractTxGenerate()
		},
	}

	t.addFlags()
	return t.cmd
}

func (t *ContractCommand) addFlags() {
	t.cmd.Flags().StringVarP(&envCfgPath, "config", "c", "./conf/generate/counter.yaml", "conf file")
}

func (t *ContractCommand) ContractTxGenerate() error {
	cfg, err := LoadConfig(envCfgPath)
	if err != nil {
		return err
	}
	t.config = cfg
	ctx := context.TODO()
	if t.config.Process == 1 {
		return t.generate(ctx)
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < t.config.Process; i++ {
		wg.Add(1)
		t.spawn(wg, i)
	}
	wg.Wait()
	return nil
}

func (t *ContractCommand) spawn(wg *sync.WaitGroup, child int) error {
	cmd := exec.Command(os.Args[0],
		"contract",
		"--config", envCfgPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() {
		defer wg.Done()
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (t *ContractCommand) generate(ctx context.Context) error {
	config := &cases.Config{
		Host:        t.config.Host,
		Total:       t.config.Total,
		Concurrency: t.config.Concurrency,
		Args:        t.config.Tags,
	}

	generator, err := cases.NewContract(config)
	if err != nil {
		return fmt.Errorf("new contract error: %v", err)
	}

	if err = generator.Init(); err != nil {
		return fmt.Errorf("init contract error: %v", err)
	}

	encoders := make([]*json.Encoder, t.config.Concurrency)
	for i := 0; i < t.config.Concurrency; i++ {
		filename := fmt.Sprintf("contract.dat.%04d", t.config.Child*t.config.Concurrency+i)
		file, err := os.Create(filepath.Join(t.config.Output, filename))
		if err != nil {
			return fmt.Errorf("open output file error: %v", err)
		}
		encoders[i] = json.NewEncoder(file)
	}

	// 生成数据1.1倍冗余
	total := int(float32(t.config.Total/t.config.Concurrency) * 1.1)
	Consumer(total, t.config.Concurrency, generator, func(i int, tx proto.Message) error {
		if err := encoders[i].Encode(tx); err != nil {
			log.Fatalf("write tx error: %v", err)
			return err
		}
		return nil
	})

	log.Printf("child=%d, pid=%d", t.config.Child, os.Getpid())
	return nil
}

func init() {
	AddCommand(NewContractCommand)
	rand.Seed(time.Now().UnixNano())
}

// 加载配置
func LoadConfig(evnPath string) (*Config, error) {
	if len(evnPath) == 0 {
		return nil, errors.New("evnPath is empty, please set evnPath")
	}
	// 读取配置文件
	configContent, err := ioutil.ReadFile(evnPath)
	if err != nil {
		return nil, errors.New("failed to read evnPath file:" + err.Error())
	}
	// 反序列化配置文件到cfg
	cfg := new(Config)
	err = yaml.Unmarshal(configContent, &cfg)
	if err != nil {
		return nil, errors.New("failed to unmarshal XConfig:" + err.Error())
	}

	return cfg, nil
}
