package monitor

import (
	"fmt"
	"os"
	"sync"
	"time"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
	"github.com/apcera/termtables"
	"github.com/struCoder/pidusage"
)

type MonInfo struct {
	Type    string
	Name    string
	MemMax  string
	MemAvg  string
	CPUMax  string
	CPUAvg  string
	MsgType common.MsgType
}

var (
	ResUsageList []MonInfo
)

func StartMonitor(done chan bool, wg *sync.WaitGroup) {
	stat := NewInfo()

	defer func() {
		InfoPrintFormat(stat)
		wg.Done()
	}()

loop:
	for {
		select {
		case <-done:
			break loop
		default:
			sysInfo, err := pidusage.GetStat(os.Getpid())
			if err != nil {
				log.ERROR.Printf("encount error <%s>", err)
				continue
			}
			stat.Add(sysInfo)

			time.Sleep(1 * time.Second)
		}
	}
}

func InfoPrintFormat(stat *StatInfo) {
	table := termtables.CreateTable()
	table.AddHeaders("Type", "Name", "Mem(max)", "Mem(avg)", "CPU(max)", "CPU(avg)")
	table.AddRow("Process", "", stat.mem.max, stat.mem.total/float64(stat.count),
		fmt.Sprintf("%.2f%%", stat.cpu.max), fmt.Sprintf("%.2f%%", stat.cpu.total/float64(stat.count)))

	monInfo := MonInfo{
		Type:    "Process",
		Name:    "",
		MemMax:  fmt.Sprintf("%.0f B", stat.mem.max),
		MemAvg:  fmt.Sprintf("%.0f B", stat.mem.total/float64(stat.count)),
		CPUMax:  fmt.Sprintf("%.2f%%", stat.cpu.max),
		CPUAvg:  fmt.Sprintf("%.2f%%", stat.cpu.total/float64(stat.count)),
		MsgType: common.MonMsg,
	}

	ResUsageList = append(ResUsageList, monInfo)
	log.INFO.Println("resource usage.....\n" + table.Render())
}
