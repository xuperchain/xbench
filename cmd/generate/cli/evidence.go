package cli

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/xuperchain/xuper-sdk-go/account"
    pb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
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

    total       int64
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
    t.cmd.Flags().Int64VarP(&t.total, "total", "t", 1000000, "total tx number")
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
        "--total", strconv.FormatInt(t.total/int64(t.process), 10),
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
    wg := new(sync.WaitGroup)
    for i := 0; i < t.concurrency; i++ {
        wg.Add(1)
        go func() {
            t.worker(file, int(t.total)/t.concurrency)
            wg.Done()
        }()
    }
    wg.Wait()

    log.Printf("child=%d, output=%s", t.child, file.Name())
    return nil
}

func (t *EvidenceCommand) worker(out io.Writer, n int) {
    txs := make([]*pb.Transaction, t.batch)
    for i := 0; i < n; i += t.batch {
        for j := 0; j < t.batch; j++ {
            txs[j] = EvidenceTx(AK, t.length)
        }
        store(txs, out)

        if i%100000 == 0 {
            log.Printf("pid=%d, count=%d\n", os.Getpid(), i)
        }
    }
}

// 存储交易：json格式
func store(txs []*pb.Transaction, out io.Writer) {
    e := json.NewEncoder(out)
    for _, tx := range txs {
        err := e.Encode(tx)
        if err != nil {
            log.Printf("store: write tx error: %v", err)
            return
        }
    }
}

func EvidenceTx(ak *account.Account, length int) *pb.Transaction {
    tx := &pb.Transaction{
        Version:   3,
        Desc:      RandBytes(length),
        Nonce:     strconv.FormatInt(time.Now().UnixNano(), 36),
        Timestamp: time.Now().UnixNano(),
        Initiator: ak.Address,
    }

    SignTx(tx, ak)
    return tx
}

func init() {
    AddCommand(NewEvidenceCommand)
    rand.Seed(time.Now().UnixNano())
}
