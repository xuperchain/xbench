package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"time"
	"github.com/xuperchain/xuperbench/config"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperbench/monitor"
)

// Report generate the html report based on the tpsinfo and resource usage info.
func Report(conf *config.Config, tpsList []monitor.TpsInfo, resUsageList []monitor.MonInfo) {
	tmpl, err := template.ParseFiles("conf/report.tpl")
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return
	}

	fn := fmt.Sprintf("report_%v.html", time.Now().Format("2006-01-02T150405.0000"))
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return
	}

	var (
		confVal    []byte
		blockChain string
	)

	if conf == nil {
		confVal, err = json.MarshalIndent(nil, "", "    ")
	} else {
		confVal, err = json.MarshalIndent(*conf, "", "    ")
	}
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return
	}

	// tpsList := stat.TpsSummary(gStat)
	if conf == nil {
		blockChain = ""
	} else {
		blockChain = string(conf.BCType)
	}
	data := struct {
		BlockChain string
		TestRound  int
		Version    string
		Config     string
		TpsList    []monitor.TpsInfo
		ResList    []monitor.MonInfo
	}{
		BlockChain: blockChain,
		TestRound:  len(tpsList),
		Version:    "V1.1.5",
		Config:     string(confVal),
		TpsList:    tpsList,
		ResList:    resUsageList,
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		log.ERROR.Printf("encount error <%s>", err)
		return
	}
	log.INFO.Printf("generate benchmark report <%s> successfully.", fn)
}
