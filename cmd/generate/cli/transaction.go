package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xuperchain/xbench/generate"
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
type TransactionCommand struct {
    cli *Cli
    cmd *cobra.Command

    total int
    split int

    input  string
    output string

    // 进程数
    process     int
    // 进程编号
    child       int
    concurrency int

    txid string
    address string
    amount string
}

type Config struct {
	Total int
	Split int
	Concurrency int

	Suffix string
}

func NewTransactionCommand(cli *Cli) *cobra.Command {
    t := new(TransactionCommand)
    t.cli = cli
    t.cmd = &cobra.Command{
        Use:   "tx",
        Short: "transaction",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := context.TODO()
            if t.process > 1 {
                return t.multiGenerate(ctx)
            }

            config := &Config{
                Total: t.total,
                Split: t.split,
                Concurrency: t.concurrency,
                Suffix: fmt.Sprintf(".child%d", t.child),
            }
            _, err := t.generate(ctx, config)
            return err
        },
    }
    t.addFlags()
    return t.cmd
}

func (t *TransactionCommand) addFlags() {
    t.cmd.Flags().IntVarP(&t.total, "total", "t", 1000000, "total tx number")
    t.cmd.Flags().IntVarP(&t.split, "split", "s", 100, "split tx output number")

    // input file
    t.cmd.Flags().StringVarP(&t.input, "input", "i", "../data/boot.json", "boot tx input path")

    // input txid
    t.cmd.Flags().StringVarP(&t.txid, "txid", "", "", "txid")
    t.cmd.Flags().StringVarP(&t.address, "address", "", "dw3RjnTe47G4u6a6hHWCfEhtaDkgdYWTE", "address")
    t.cmd.Flags().StringVarP(&t.amount, "amount", "", "1000000000000", "amount")

    t.cmd.Flags().StringVarP(&t.output, "output", "o", "../data/transaction", "generate tx output path")
    t.cmd.Flags().IntVarP(&t.process, "process", "", 1, "process number")
    t.cmd.Flags().IntVarP(&t.child, "child", "", 0, "child number")
    t.cmd.Flags().IntVarP(&t.concurrency, "concurrency", "c", 20, "goroutine concurrency number")
}

func (t *TransactionCommand) multiGenerate(ctx context.Context) error {
    config := &Config{
        Total: t.process,
        Split: t.process,
        Concurrency: t.process,
	    Suffix: ".parent",
    }
    childTxFile, err := t.generate(ctx, config)
    if err != nil {
        return err
    }

    log.Printf("process=%d, output=%s", t.process, childTxFile)

    wg := new(sync.WaitGroup)
    for child := 0; child < t.process; child++ {
        wg.Add(1)
        t.spawn(wg, childTxFile, child)
    }
    wg.Wait()
    return nil
}

func (t *TransactionCommand) spawn(wg *sync.WaitGroup, input string, child int) error {
    cmd := exec.Command(os.Args[0],
        "tx",
        "--total", strconv.FormatInt(int64(t.total/t.process), 10),
        "--split", strconv.Itoa(t.split),
        "--input", input,
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

func (t *TransactionCommand) generate(ctx context.Context, config *Config) (string, error) {
    accounts, err := generate.LoadAccount(config.Split)
    if err != nil {
        return "", fmt.Errorf("load account error: %v", err)
    }

    var txs []*pb.Transaction
    if len(t.txid) > 0 {
        tx, err := generate.BootTx(t.txid, t.address, t.amount)
        if err != nil {
            return "", err
        }
        txs = append(txs, tx)
    } else {
        txs, err = ReadTxs(t.input)
        if err != nil {
            return "", fmt.Errorf("read boot tx error: %v", err)
        }
    }

    if len(txs) < 1 {
        return "", fmt.Errorf("boot tx not exist")
    }

	log.Printf("boot=%s\n", generate.FormatTx(txs[t.child]))
	transaction, err := generate.NewTransaction(config.Total, config.Concurrency, config.Split, accounts, txs[t.child])
	if err != nil {
		return "", fmt.Errorf("new evidence error: %v", err)
	}

	level := transaction.Level()
	levelFile := make([]*os.File, level+1)
	for i := 0; i <= level; i++ {
		filename := fmt.Sprintf("%02d.dat%s", i, config.Suffix)
		file, err := os.Create(filepath.Join(t.output, filename))
		if err != nil {
			return "", fmt.Errorf("open level file error: %v", err)
		}
		levelFile[i] = file
	}

	transaction.Consumer(func(level int, txs []*pb.Transaction) error {
		if err := store(txs, levelFile[level]); err != nil {
			log.Printf("store: write tx error: %v", err)
			return err
		}
		return nil
	})

	filename := levelFile[level].Name()
    log.Printf("child=%d, output=%s", t.child, filename)
    return filename, nil
}

func ReadTxs(input string) ([]*pb.Transaction, error) {
    file, err := os.Open(input)
    if err != nil {
        log.Printf("open tx file error: %v", err)
        return nil, fmt.Errorf("open input file error: %v", err)
    }

    d := json.NewDecoder(file)

    txs := make([]*pb.Transaction, 0, 16)
    for {
        tx := &pb.Transaction{}
        err = d.Decode(tx)
        if err != nil {
            if err == io.EOF {
                err = nil
            }
            break
        }

        txs = append(txs, tx)
    }

    return txs, err
}

func init() {
    AddCommand(NewTransactionCommand)
    rand.Seed(time.Now().UnixNano())
}
