package monitor

import (
	"encoding/json"
	"fmt"
	"time"
	"github.com/xuperchain/xuperbench/log"
	"github.com/apcera/termtables"
)

type TpsInfo struct {
	Round    int
	Name     string
	Succ     int
	Fail     int
	SendRate string
	MaxLat   time.Duration
	MinLat   time.Duration
	AvgLat   time.Duration
	Tps      string
}

func TpsSummary(testStats []*LatencyStats) []TpsInfo {
	ret := make([]TpsInfo, 0, len(testStats))
	for i, v := range testStats {
		tps := TpsInfo{
			Round:    i + 1,
			Name:     v.ext,
			Succ:     v.succ,
			Fail:     v.fail,
			SendRate: fmt.Sprintf("%d", v.RatePerSec()),
			MaxLat:   v.max,
			MinLat:   v.min,
			AvgLat:   v.Avg(),
			Tps:      fmt.Sprintf("%v tps", v.TPS()),
		}
		ret = append(ret, tps)
	}

	return ret
}

func PrintFormat(testStats []*LatencyStats) {
	table := termtables.CreateTable()

	table.AddHeaders("Round", "Name", "Succ", "Fail", "Send Rate",
		"Max Latency", "Min Latency", "Avg Latency", "Throughput")
	for i, v := range testStats {
		table.AddRow(i+1, v.ext, v.succ, v.fail, fmt.Sprintf("%d", v.RatePerSec()),
			v.max, v.min, v.Avg(), fmt.Sprintf("%v tps", v.TPS()))
	}

	log.INFO.Println("benchmark summary...\n" + table.Render())

	ptable := termtables.CreateTable()
	ptable.AddHeaders("Round", "50%", "60%", "70%", "80%", "90%", "95%", "97%", "99%")
	for i, v := range testStats {
		ptable.AddRow(i+1, v.Pct(50), v.Pct(60), v.Pct(70), v.Pct(80), v.Pct(90),
			v.Pct(95), v.Pct(97), v.Pct(99))
	}

	log.INFO.Println("benchmark percentile summary...\n" + ptable.Render())
}

func Marshal(gStat []*LatencyStats) string {
	ret := make([]map[string]interface{}, 0)
	for i, v := range gStat {
		m := make(map[string]interface{})
		m["RoundId"] = i + 1
		m["Name"] = v.ext
		m["Succ"] = v.succ
		m["Fail"] = v.fail
		m["MaxLat"] = v.max
		m["MinLat"] = v.min
		m["AvgLat"] = v.Avg()
		m["TotLat"] = v.dur
		m["Begin"] = v.begin.UnixNano()
		m["End"] = v.end.UnixNano()
		m["Sendrate"] = v.RatePerSec()

		ret = append(ret, m)
	}

	data, err := json.Marshal(ret)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return ""
	}

	return string(data)
}
