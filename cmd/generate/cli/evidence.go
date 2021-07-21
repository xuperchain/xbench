package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xbench/cases"
	"github.com/xuperchain/xuperchain/service/pb"
)

// BenchCommand
type EvidenceCommand struct {
	cli *Cli
	cmd *cobra.Command

	// 交易总量
	total int
	// 并发数
	concurrency int
	// 产出路径
	output string

	// 进程数
	process     int
	// 进程编号
	child       int

	// 存证大小
	length      int
}

func NewEvidenceCommand(cli *Cli) *cobra.Command {
	t := new(EvidenceCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "evidence",
		Short: "evidence",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			if t.process == 1 {
				return t.generate(ctx)
			}

			wg := new(sync.WaitGroup)
			for i := 0; i < t.process; i++ {
				wg.Add(1)
				t.spawn(wg, i)
			}
			wg.Wait()
			return nil
		},
	}
	t.addFlags()
	return t.cmd
}

func (t *EvidenceCommand) addFlags() {
	t.cmd.Flags().IntVarP(&t.total, "total", "t", 1000000, "total tx number")
	t.cmd.Flags().IntVarP(&t.concurrency, "concurrency", "c", 20, "goroutine concurrency number")
	t.cmd.Flags().StringVarP(&t.output, "output", "o", "./data/evidence", "generate tx output path")
	t.cmd.Flags().IntVarP(&t.length, "length", "l", 200, "evidence data length")

	t.cmd.Flags().IntVarP(&t.process, "process", "", 1, "process number")
	t.cmd.Flags().IntVarP(&t.child, "child", "", 0, "child number")
}

func (t *EvidenceCommand) spawn(wg *sync.WaitGroup, child int) error {
	cmd := exec.Command(os.Args[0],
		"evidence",
		"--total", strconv.FormatInt(int64(t.total/t.process), 10),
		"--length", strconv.Itoa(t.length),
		"--output", t.output,
		"--concurrency", strconv.Itoa(t.concurrency),
		"--process", "1",
		"--child", strconv.Itoa(child),
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


func (t *EvidenceCommand) generate(ctx context.Context) error {
	config := &cases.Config{
		Total: t.total,
		Concurrency: t.concurrency,
		Args: map[string]string{
			"length": strconv.Itoa(t.length),
		},
	}
	generator, err := cases.NewEvidence(config)
	if err != nil {
		return fmt.Errorf("new evidence error: %v", err)
	}

	if err = generator.Init(); err != nil {
		return fmt.Errorf("init evidence error: %v", err)
	}

	encoders := make([]*json.Encoder, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		filename := fmt.Sprintf("evidence.dat.%04d", t.child*t.concurrency+i)
		file, err := os.Create(filepath.Join(t.output, filename))
		if err != nil {
			return fmt.Errorf("open output file error: %v", err)
		}
		encoders[i] = json.NewEncoder(file)
	}

	// 生成数据1.1倍冗余
	total := int(float32(t.total/t.concurrency)*1.1)
	cases.Consumer(total, t.concurrency, generator, func(i int, tx *pb.Transaction) error {
		if err := encoders[i].Encode(tx); err != nil {
			log.Fatalf("write tx error: %v", err)
			return err
		}
		return nil
	})

	log.Printf("child=%d, pid=%d", t.child, os.Getpid())
	return nil
}

func init() {
	AddCommand(NewEvidenceCommand)
	rand.Seed(time.Now().UnixNano())
}
