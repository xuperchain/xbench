package benchmark

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/config"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperbench/monitor"
	"github.com/xuperchain/xuperbench/report"
	"github.com/apcera/termtables"
	"github.com/gomodule/redigo/redis"
)

// RedisPush push msg to the broker.
func RedisPush(conf *config.Config, data []string) {
	c, err := redis.Dial("tcp", conf.Broker)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return
	}

	for _, v := range data {
		_, err = c.Do("rpush", conf.ResultBackend, v)
		if err != nil {
			log.ERROR.Printf("encount error <%s>", err)
		}
	}
}

func getBackendResult(conf *config.Config) []string {
	results := make([]string, 0)
	c, err := redis.Dial("tcp", conf.Broker)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return nil
	}

	log.INFO.Println("wait remote client to finish.....")
	msg, err := c.Do("blpop", conf.ResultBackend, 0)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return nil
	}
	log.DEBUG.Printf("recive msg: <%s>", string(msg.([]interface{})[1].([]byte)))
	results = append(results, string(msg.([]interface{})[1].([]byte)))
	for {
		msg, err := c.Do("blpop", conf.ResultBackend, 10)
		if err != nil {
			log.ERROR.Printf("encount error <%s>", err)
			continue
		}
		if msg == nil {
			break
		}

		log.DEBUG.Printf("recive msg: <%s>", string(msg.([]interface{})[1].([]byte)))
		results = append(results, string(msg.([]interface{})[1].([]byte)))
	}

	return results
}

// BackendProf get tpsinfo msg from broker and then alaylis.
func BackendProf(conf *config.Config) {
	data := getBackendResult(conf)
	var (
		tpsData      []string
		resUsageData []string
		tv           []map[string]interface{}
	)

	for _, v := range data {
		match, _ := regexp.MatchString(string(common.MonMsg), v)
		if match {
			resUsageData = append(resUsageData, v)
		} else {
			tpsData = append(tpsData, v)
		}
	}

	max := func(x, y float64) float64 {
		if x > y {
			return x
		}
		return y
	}

	min := func(x, y float64) float64 {
		if x < y {
			return x
		}
		return y
	}

	for _, v := range tpsData {
		var stat []map[string]interface{}

		json.Unmarshal([]byte(v), &stat)
		if len(tv) == 0 {
			tv = stat
		} else {
			for ii, vv := range stat {
				tv[ii]["Succ"] = tv[ii]["Succ"].(float64) + vv["Succ"].(float64)
				tv[ii]["Fail"] = tv[ii]["Fail"].(float64) + vv["Fail"].(float64)
				tv[ii]["MaxLat"] = max(tv[ii]["MaxLat"].(float64), vv["MaxLat"].(float64))
				tv[ii]["MinLat"] = min(tv[ii]["MinLat"].(float64), vv["MinLat"].(float64))
				tv[ii]["AvgLat"] = (tv[ii]["AvgLat"].(float64) + vv["AvgLat"].(float64)) / 2
				tv[ii]["TotLat"] = tv[ii]["TotLat"].(float64) + vv["TotLat"].(float64)
				tv[ii]["Begin"] = min(tv[ii]["Begin"].(float64), vv["Begin"].(float64))
				tv[ii]["End"] = max(tv[ii]["End"].(float64), vv["End"].(float64))
				tv[ii]["Sendrate"] = tv[ii]["Sendrate"].(float64) + vv["Sendrate"].(float64)
			}
		}
	}

	tpsList := make([]monitor.TpsInfo, 0, len(tv))
	resUsageList := make([]monitor.MonInfo, 0, len(resUsageData))

	table := termtables.CreateTable()

	table.AddHeaders("Round", "Name", "Succ", "Fail", "Send Rate",
		"Max Latency", "Min Latency", "Avg Latency", "Throughput")
	for _, v := range tv {
		lat := (v["End"].(float64) - v["Begin"].(float64)) / float64(time.Second)
		tps := fmt.Sprintf("%v tps", int(v["Succ"].(float64)/lat))
		table.AddRow(
			int(v["RoundId"].(float64)), v["Name"],
			int(v["Succ"].(float64)), int(v["Fail"].(float64)), "TODO",
			time.Duration(v["MaxLat"].(float64)),
			time.Duration(v["MinLat"].(float64)),
			time.Duration(v["AvgLat"].(float64)), tps,
		)

		tpsInfo := monitor.TpsInfo{
			Round:    int(v["RoundId"].(float64)),
			Name:     v["Name"].(string),
			Succ:     int(v["Succ"].(float64)),
			Fail:     int(v["Fail"].(float64)),
			SendRate: v["Sendrate"].(string),
			MaxLat:   time.Duration(v["MaxLat"].(float64)),
			MinLat:   time.Duration(v["MinLat"].(float64)),
			AvgLat:   time.Duration(v["AvgLat"].(float64)),
			Tps:      tps,
		}
		tpsList = append(tpsList, tpsInfo)
	}

	for _, v := range resUsageData {
		var monInfo monitor.MonInfo
		json.Unmarshal([]byte(v), &monInfo)

		log.DEBUG.Printf("%#v", monInfo)
		resUsageList = append(resUsageList, monInfo)
	}

	log.INFO.Println("benchmark total summary.....\n", table.Render())
	report.Report(conf, tpsList, resUsageList)
}
