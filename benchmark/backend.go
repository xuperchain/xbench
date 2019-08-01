package benchmark

import (
	"regexp"
	"strconv"
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/config"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperbench/monitor"
	"github.com/xuperchain/xuperbench/report"
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

func RedisClear(conf *config.Config) {
	c, err := redis.Dial("tcp", conf.Broker)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return
	}
	prefix := conf.PubSubChan + "_"
	_, err = c.Do("del", prefix + "worker")
	rnd := 0
	for {
		rsp, _ := c.Do("del", prefix + "rnd_" + strconv.Itoa(rnd))
		if rsp.(int64) != int64(1) {
			break
		}
		rnd += 1
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

		results = append(results, string(msg.([]interface{})[1].([]byte)))
	}

	return results
}

// BackendProf get tpsinfo msg from broker and then alaylis.
func BackendProf(conf *config.Config) {
	data := getBackendResult(conf)
	RedisClear(conf)
	var (
		tpsData      []string
		resUsageData []string
		tv           []*monitor.LatencyStats
	)

	for _, v := range data {
		match, _ := regexp.MatchString(string(common.MonMsg), v)
		if match {
			resUsageData = append(resUsageData, v)
		} else {
			tpsData = append(tpsData, v)
		}
	}

	for _, v := range tpsData {
		stat := monitor.Deserialize(v)
		if len(tv) == 0 {
			for _, vv := range stat {
				itm := monitor.New(vv.Ext)
				itm.Add(&vv)
				tv = append(tv, itm)
			}
		} else {
			for ii, vv := range stat {
				tv[ii].Add(&vv)
			}
		}
	}

	tpsList := monitor.TpsSummary(tv)
	resUsageList := make([]monitor.MonInfo, 0, len(resUsageData))

	for _, v := range resUsageData {
		var monInfo monitor.MonInfo
		json.Unmarshal([]byte(v), &monInfo)

		log.DEBUG.Printf("%#v", monInfo)
		resUsageList = append(resUsageList, monInfo)
	}
	monitor.PrintFormat(tv)
	report.Report(conf, tpsList, resUsageList)
}
