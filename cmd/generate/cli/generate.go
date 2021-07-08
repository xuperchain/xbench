package cli

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/xuperchain/xuper-sdk-go/account"
    pb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
    "github.com/xuperchain/xupercore/protos"
    "log"
    "math"
    "math/big"
    "os"
    "path/filepath"
    "strconv"
    "sync"
    "sync/atomic"
    "time"
)

type Level struct {
    txs []*pb.Transaction
    level int
}

type Config struct {
    Total int64
    Split int
    Concurrency int

    Path string
    BaseLevel int
}

type generator struct {
    ctx context.Context
    config *Config

    accounts    []*account.Account
    addresses   map[string]*account.Account

    level       int
    levelCh     chan Level
    levelFile   []*os.File
}

type Generator interface {
    Generate(tx *pb.Transaction)
    String() string
}

// total 与 split 为指数关系
func NewGenerator(config *Config, accounts []*account.Account) (Generator, error) {
    if config.Total <= 1 {
        return nil, fmt.Errorf("split error: split=%v", config.Total)
    }

    if len(accounts) < config.Split {
        return nil, fmt.Errorf("accounts not enough: split=%d, accounts=%d", config.Split, len(accounts))
    }

    addresses := make(map[string]*account.Account, len(accounts))
    for _, ak := range accounts {
        addresses[ak.Address] = ak
    }

    t := &generator{
        config: config,
        accounts: accounts,
        addresses: addresses,
        levelCh: make(chan Level, 10*config.Split),
    }

    level := math.Log10(float64(config.Total)) / math.Log10(float64(config.Split))
    t.level = int(math.Ceil(level))
    log.Printf("generator: total=%d split=%d level=%d", config.Total, config.Split, t.level)

    t.levelFile = make([]*os.File, t.level+1)
    for i := 0; i <= t.level; i++ {
        filename := fmt.Sprintf("%02d.dat.%d", i, os.Getpid())
        file, err := os.Create(filepath.Join(t.config.Path, filename))
        if err != nil {
            return nil, fmt.Errorf("open level file error: %v", err)
        }
        t.levelFile[i] = file
    }

    ctx, cancel := context.WithCancel(context.Background())
    t.ctx = ctx

    go t.output(cancel)
    return t, nil
}

func (t *generator) String() string {
    return t.levelFile[len(t.levelFile)-1].Name()
}

// 生产交易
func (t *generator) Generate(tx *pb.Transaction) {
    t.generate(tx, 0)

    select {
    case <-t.ctx.Done():
        return
    }
}

func (t *generator) generate(tx *pb.Transaction, level int) {
    if level >= t.level {
        return
    }

    select {
    case <-t.ctx.Done():
        return
    default:
    }

    unconfirmedTxs := t.producer(tx, t.config.Split)
    t.levelCh <- Level{
        txs: unconfirmedTxs,
        level: level,
    }
    for _, unconfirmedTx := range unconfirmedTxs {
        t.generate(unconfirmedTx, level+1)
    }
}

// 产出交易：level之间是关联交易，level内是无关联交易
func (t *generator) output1(cancel context.CancelFunc) {
    var count int64
    for {
        select {
        case txsLevel := <-t.levelCh:
            t.store(txsLevel.txs, txsLevel.level)
            if txsLevel.level+1 < t.level {
                continue
            }

            // 叶子交易
            for _, tx := range txsLevel.txs {
                txs := t.producer(tx, 1)
                t.store(txs, txsLevel.level+1)

                count += int64(len(txs))
                if count%100000 == 0 {
                    log.Printf("pid=%d, count=%d\n", os.Getpid(), count)
                }

                if count >= t.config.Total {
                    cancel()
                    return
                }
            }
        }
    }
}

// 产出交易：level之间是关联交易，level内是无关联交易
func (t *generator) output(cancel context.CancelFunc) {
    var count int64
    for {
        select {
        case txsLevel := <-t.levelCh:
            t.store(txsLevel.txs, txsLevel.level)
            if txsLevel.level+1 < t.level {
                continue
            }

            ch := make(chan struct{}, t.config.Concurrency)
            for i := 0; i < t.config.Concurrency; i++ {
                ch <- struct{}{}
            }

            var wg sync.WaitGroup
            for _, tx := range txsLevel.txs {
                wg.Add(1)
                <- ch
                go func(tx *pb.Transaction) {
                    txs := t.producer(tx, 1)
                    t.store(txs, txsLevel.level+1)

                    total := atomic.AddInt64(&count, int64(len(txs)))
                    if total%100000 == 0 {
                        log.Printf("count=%d\n", count)
                    }

                    ch <- struct {}{}
                    wg.Done()
                }(tx)

            }
            wg.Wait()

            if count >= t.config.Total {
                cancel()
                return
            }
        }
    }
}

// 存储交易：json格式
func (t *generator) store(txs []*pb.Transaction, level int) {
    e := json.NewEncoder(t.levelFile[level])
    for _, tx := range txs {
        err := e.Encode(tx)
        if err != nil {
            log.Printf("store: write tx error: %v", err)
            return
        }
    }
}

// 生成子交易
func (t *generator) producer(tx *pb.Transaction, split int) []*pb.Transaction {
    txs := make([]*pb.Transaction, len(tx.TxOutputs))
    for i, txOutput := range tx.TxOutputs {
        if txOutput == nil {
            continue
        }

        ak := t.addresses[string(txOutput.ToAddr)]
        if ak == nil {
            ak = AK
        }

        input := &protos.TxInput{
            RefTxid: tx.Txid,
            RefOffset: int32(i),
            FromAddr: txOutput.ToAddr,
            Amount: txOutput.Amount,
        }
        output := &protos.TxOutput{
            ToAddr: []byte(t.accounts[i].Address),
            Amount: txOutput.Amount,
        }
        newTx := TransferTx(input, output, split)
        txs[i] = SignTx(newTx, ak)
    }

    return txs
}

func TransferTx(input *protos.TxInput, output *protos.TxOutput, split int) *pb.Transaction {
    tx := &pb.Transaction{
        Version:   3,
        //Desc:      RandBytes(200),
        Nonce:     strconv.FormatInt(time.Now().UnixNano(), 36),
        Timestamp: time.Now().UnixNano(),
        Initiator: string(input.FromAddr),
        TxInputs: []*protos.TxInput{input},
        TxOutputs: Split(output, split),
    }

    return tx
}

func Split(txOutput *protos.TxOutput, split int) []*protos.TxOutput {
    if split <= 1 {
        return []*protos.TxOutput{txOutput}
    }

    total := big.NewInt(0).SetBytes(txOutput.Amount)
    if big.NewInt(int64(split)).Cmp(total) == 1 {
        // log.Printf("split utxo <= balance required")
        panic("amount not enough")
        return []*protos.TxOutput{txOutput}
    }

    amount := big.NewInt(0)
    amount.Div(total, big.NewInt(int64(split)))

    output := protos.TxOutput{}
    output.Amount = amount.Bytes()
    output.ToAddr = txOutput.ToAddr

    rest := total
    txOutputs := make([]*protos.TxOutput, 0, split+1)
    for i := 1; i < split && rest.Cmp(amount) == 1; i++ {
        tmpOutput := output
        txOutputs = append(txOutputs, &tmpOutput)
        rest.Sub(rest, amount)
    }
    output.Amount = rest.Bytes()
    txOutputs = append(txOutputs, &output)

    return txOutputs
}
