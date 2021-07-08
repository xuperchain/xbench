package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	pb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	"github.com/xuperchain/xupercore/lib/utils"
	"github.com/xuperchain/xupercore/protos"
	"io"
	"log"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// BenchCommand
type TransactionCommand struct {
    cli *Cli
    cmd *cobra.Command

    total int64
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
                Path: t.output,
                ID: t.child+1,
            }
            _, err := t.generate(ctx, config)
            return err
        },
    }
    t.addFlags()
    return t.cmd
}

func (t *TransactionCommand) addFlags() {
    t.cmd.Flags().Int64VarP(&t.total, "total", "t", 1000000, "total tx number")
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
        Total: int64(t.process),
        Split: t.process,
        Concurrency: t.process,
        //Path: filepath.Join(t.output, "parent"),
	    Path: t.output,
	    ID: 0,
    }
    g, err := t.generate(ctx, config)
    if err != nil {
        return err
    }

    childTxFile := g.String()
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
        "--total", strconv.FormatInt(t.total/int64(t.process), 10),
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

func (t *TransactionCommand) generate(ctx context.Context, config *Config) (Generator, error) {
    accounts, err := LoadAccount(t.split)
    if err != nil {
        return nil, fmt.Errorf("load account error: %v", err)
    }

    g, err := NewGenerator(config, accounts)
    if err != nil {
        return nil, fmt.Errorf("new generator error: %v", err)
    }

    var txs []*pb.Transaction
    if len(t.txid) > 0 {
        tx, err := t.BootTx()
        if err != nil {
            return nil, err
        }
        txs = append(txs, tx)
    } else {
        txs, err = ReadTxs(t.input)
        if err != nil {
            return nil, fmt.Errorf("read boot tx error: %v", err)
        }
    }

    if len(txs) < 1 {
        return nil, fmt.Errorf("boot tx not exist")
    }

    log.Printf("boot=%s\n", FormatTx(txs[t.child]))
    g.Generate(txs[t.child])
    log.Printf("child=%d, output=%s", t.child, g.String())
    return g, nil
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

func (t *TransactionCommand) BootTx() (*pb.Transaction, error) {
    if len(t.txid) <= 0 {
        return nil, fmt.Errorf("txid not exist")
    }

    amount, ok := big.NewInt(0).SetString(t.amount, 10)
    if !ok {
        return nil, fmt.Errorf("boot tx amount error")
    }

    tx := &pb.Transaction{
        Txid: utils.DecodeId(t.txid),
        TxOutputs: []*protos.TxOutput{
            {
                ToAddr: []byte(t.address),
                Amount: amount.Bytes(),
            },
        },
    }

    return tx, nil
}

func init() {
    AddCommand(NewTransactionCommand)
    rand.Seed(time.Now().UnixNano())
}
