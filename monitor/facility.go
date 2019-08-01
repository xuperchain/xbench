package monitor

import (
	"encoding/json"
	"encoding/gob"
	"bytes"
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
		fmt.Println(v)
		tps := TpsInfo{
			Round:    i + 1,
			Name:     v.Ext,
			Succ:     v.Succ,
			Fail:     v.Fail,
			SendRate: fmt.Sprintf("%d", v.RatePerSec()),
			MaxLat:   v.Max,
			MinLat:   v.Min,
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
		table.AddRow(i+1, v.Ext, v.Succ, v.Fail, fmt.Sprintf("%d", v.RatePerSec()),
			v.Max, v.Min, v.Avg(), fmt.Sprintf("%v tps", v.TPS()))
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
		m["Name"] = v.Ext
		m["Succ"] = v.Succ
		m["Fail"] = v.Fail
		m["MaxLat"] = v.Max
		m["MinLat"] = v.Min
		m["AvgLat"] = v.Avg()
		m["TotLat"] = v.Dur
		m["Begin"] = v.Begin.UnixNano()
		m["End"] = v.End.UnixNano()
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

func Serialize(gStat []*LatencyStats) string {
	data := []LatencyStats{}
	for _, v := range gStat {
		data = append(data, *v)
	}
	fmt.Println(data)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return ""
	}
	return buf.String()
}

func Deserialize(str string) []LatencyStats {
	var stat []LatencyStats
	buf := bytes.NewBufferString(str)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&stat)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return nil
	}
	return stat
}
