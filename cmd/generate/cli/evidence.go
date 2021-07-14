package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xuperchain/xbench/generate"
	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/service/pb"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// BenchCommand
type EvidenceCommand struct {
    cli *Cli
    cmd *cobra.Command

    total       int
    // 存证大小
    length      int
    // 输出目录
    output      string
    // 一个批次生产交易量
    batch       int
    // 进程数
    process     int
    // 进程编号
    child       int
    concurrency int
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
    t.cmd.Flags().IntVarP(&t.length, "length", "l", 200, "evidence data length")
    t.cmd.Flags().IntVarP(&t.batch, "batch", "", 1000, "tx batch number")
    t.cmd.Flags().StringVarP(&t.output, "output", "o", "./data/evidence", "generate tx output path")
    t.cmd.Flags().IntVarP(&t.process, "process", "", 1, "process number")
    t.cmd.Flags().IntVarP(&t.child, "child", "", 0, "child number")
    t.cmd.Flags().IntVarP(&t.concurrency, "concurrency", "c", 20, "goroutine concurrency number")
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
    filename := fmt.Sprintf("%02d.dat", t.child)
    file, err := os.Create(filepath.Join(t.output, filename))
    if err != nil {
        return fmt.Errorf("open output file error: %v", err)
    }

    log.Printf("child=%d, pid=%d", t.child, os.Getpid())

    evidence, err := generate.NewEvidence(t.total, t.concurrency, t.length, t.batch)
    if err != nil {
    	return fmt.Errorf("new evidence error: %v", err)
    }

	for txs := range evidence.Generate() {
		if err := store(txs, file); err != nil {
			return fmt.Errorf("store: write tx error: %v", err)
		}
	}

    log.Printf("child=%d, pid=%d， output=%s", t.child, os.Getpid(), file.Name())
    return nil
}

// 存储交易：json格式
func store(txs []*pb.Transaction, out io.Writer) error {
    e := json.NewEncoder(out)
    for _, tx := range txs {
        err := e.Encode(tx)
        if err != nil {
            return err
        }
    }

    return nil
}

func init() {
    AddCommand(NewEvidenceCommand)
    rand.Seed(time.Now().UnixNano())
}
