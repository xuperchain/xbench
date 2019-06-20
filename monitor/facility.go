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
			SendRate: "TODO",
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
		table.AddRow(i+1, v.ext, v.succ, v.fail, "TODO",
			v.max, v.min, v.Avg(), fmt.Sprintf("%v tps", v.TPS()))
	}

	log.INFO.Println("benchmark summary...\n" + table.Render())
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

		ret = append(ret, m)
	}

	data, err := json.Marshal(ret)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return ""
	}

	return string(data)
}
