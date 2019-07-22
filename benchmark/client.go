package benchmark

import (
	"fmt"
	"sync"
	"time"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperbench/monitor"
)

// Worker do the bench test actually
func Worker(msg common.BenchMsg, wg *sync.WaitGroup, gStat chan *monitor.LatencyStats, workerSeq int) {
	testStat := monitor.New(fmt.Sprintf("%s", msg.TestCase))
	defer func() {
		gStat <- testStat
		wg.Done()
	}()

	cb := msg.CB

	runTest := func(seq int) {
		testStat.Start()
		err := cb.Run(seq, msg.Env, msg.Args)
		if err != nil {
			testStat.RecordFail(1)
		} else {
			testStat.RecordSucc(1)
		}
	}

	if msg.TxDuration > 0 {
		timeout := time.After(time.Duration(msg.TxDuration) * time.Second)
	loop:
		for {
			select {
			// TODO
			// case <-time.After(time.Duration(msg.TxDuration) * time.Second):
			case <-timeout:
				log.DEBUG.Printf("benchmark timeout %v", msg.TxDuration)
				break loop
			default:
				runTest(workerSeq)
			}
		}
	} else {
		for num := 0; num < msg.TxNumber; num++ {
			if workerSeq == 0 && num % (msg.TxNumber / 10) == 0 && num > 0 {
				log.INFO.Printf("run benchmark %d %%", num * 100 / msg.TxNumber)
			}
			runTest(workerSeq)
		}
	}
}
