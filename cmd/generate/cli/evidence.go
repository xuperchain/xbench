package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xbench/generate"
	"github.com/xuperchain/xuperchain/service/pb"
)

// BenchCommand
type EvidenceCommand struct {
    cli *Cli
    cmd *cobra.Command

    total       int
	concurrency int
    // 存证大小
    length      int
    // 输出目录
    output      string

    // 进程数
    process     int
    // 进程编号
    child       int
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
    config := &generate.Config{
    	Total: t.total,
    	Concurrency: t.concurrency,
    	Args: map[string]string{
    		"length": strconv.Itoa(t.length),
	    },
    }
    evidence, err := generate.NewEvidence(config)
    if err != nil {
    	return fmt.Errorf("new evidence error: %v", err)
    }

    queues := evidence.Generate()

	wg := new(sync.WaitGroup)
	wg.Add(len(queues))
    for i, queue := range queues {
	    filename := fmt.Sprintf("evidence.dat.%02d%02d", t.child, i)
	    file, err := os.Create(filepath.Join(t.output, filename))
	    if err != nil {
		    return fmt.Errorf("open output file error: %v", err)
	    }

    	go func(queue chan *pb.Transaction, out io.Writer) {
    		defer wg.Done()
		    if err := generate.WriteFile(queue, out); err != nil {
		    	log.Fatalf("store: write tx error: %v", err)
		    }
	    }(queue, file)
    }
    wg.Wait()

    log.Printf("child=%d, pid=%d", t.child, os.Getpid())
    return nil
}

// 自定义消费交易的方式
func Consumer(queues []chan *pb.Transaction, consume func(i int, tx *pb.Transaction) error) {
	wg := new(sync.WaitGroup)
	wg.Add(len(queues))
	dealer := func(i int, queue chan *pb.Transaction) {
		defer wg.Done()
		for txs := range queue {
			if err := consume(i, txs); err != nil {
				return
			}
		}
	}

	for i, queue := range queues {
		go dealer(i, queue)
	}
	wg.Wait()
}

func BatchWriteFile(queue chan []*pb.Transaction, out io.Writer) error {
	e := json.NewEncoder(out)
	for txs := range queue {
		for _, tx := range txs {
			err := e.Encode(tx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
    AddCommand(NewEvidenceCommand)
    rand.Seed(time.Now().UnixNano())
}
