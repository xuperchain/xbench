package benchmark

import (
	"encoding/json"
	"runtime"
	"sync"
	"strconv"
	"time"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/config"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperbench/monitor"
	"github.com/xuperchain/xuperbench/report"
)

func init() {
	cpuNum := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuNum)
}

// BenchRun control the bench process based on the config.
func BenchRun(conf *config.Config) {
	benchMsgs := config.GetBenchMsgFromConf(conf)

	log.INFO.Printf("begin to benchmark.....")
	if conf.Mode == common.RemoteMode {
		Set(conf, "worker")
	}

	gStat := make([]*monitor.LatencyStats, 0, len(benchMsgs))

	for i, testMsg := range benchMsgs {
		monDone := make(chan bool)
		monWg := new(sync.WaitGroup)
		monWg.Add(1)
		go monitor.StartMonitor(monDone, monWg)

		parallel := testMsg.Parallel

		cb := testMsg.CB
		cb.Init(parallel, testMsg.Env)

		if conf.Mode == common.RemoteMode {
			Set(conf, "rnd_" + strconv.Itoa(i))
			Wait(conf, "rnd_" + strconv.Itoa(i))
		}

		wg := new(sync.WaitGroup)
		wg.Add(parallel)
		roundStat := make(chan *monitor.LatencyStats, parallel)
		for ii := 0; ii < parallel; ii++ {
			go Worker(*testMsg, wg, roundStat, ii)
		}

		wg.Wait()
		cb.End()

		close(roundStat)
		gStat = append(gStat, monitor.Merge(roundStat))

		log.INFO.Printf("test rouond <%v> finish, wait 2s for next round!", i+1)
		time.Sleep(2 * time.Second)

		close(monDone)
		monWg.Wait()
	}

	monitor.PrintFormat(gStat)
	if conf.Mode == common.RemoteMode {
		tpsData := monitor.Serialize(gStat)
		//tpsData := monitor.Marshal(gStat)

		var (
			resUsage []byte
			err      error
		)

		i := len(monitor.ResUsageList)
		if i > 0 {
			resUsage, err = json.Marshal(monitor.ResUsageList[i-1])
			if err != nil {
				log.ERROR.Printf("encount error <%s>", err)
				resUsage = nil
			}
		}

		RedisPush(conf, []string{tpsData, string(resUsage)})
	} else {
		tpsList := monitor.TpsSummary(gStat)
		report.Report(conf, tpsList, monitor.ResUsageList)
	}
}
