package main

import (
	"flag"
	"os"
	"runtime/pprof"
	"github.com/xuperchain/xuperbench/benchmark"
	"github.com/xuperchain/xuperbench/config"
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/log"
)

var (
	configFile string
	ppf string
	worker     bool
	master     bool
)

func init() {
	flag.StringVar(&configFile, "c", "demo.json", "test config file")
	flag.StringVar(&ppf, "prof", "", "record profiling")
	flag.BoolVar(&worker, "worker", false, "benchmark worker(client)")
	flag.BoolVar(&master, "master", false, "benchmark master(send benchmsg to client)")
	flag.Parse()
}

func main() {
	if ppf != "" {
		fd, _ := os.Create(ppf)
		pprof.StartCPUProfile(fd)
		defer pprof.StopCPUProfile()
	}
	conf := config.ParseConfig(configFile)
	if conf == nil {
		log.ERROR.Printf("encount err: get %v config", conf)
		os.Exit(1)
	}
	log.DEBUG.Printf("%#v", *conf)

	if conf.Mode == common.LocalMode {
		benchmark.BenchRun(conf)
	} else if conf.Mode == common.RemoteMode {
		if worker && master {
			log.ERROR.Print("cannot both set `-worker` and `-master` flag")
			return
		}
		if !worker && !master {
			log.ERROR.Print("has not set `-worker` or `-master` flag")
			return
		}
		if worker {
			benchmark.Subscribe(conf)
		} else if master {
			benchmark.Publish(conf)
			benchmark.BackendProf(conf)
		}
	} else if conf.Mode == common.FunctionMode {
//		behavmark.RunSuite(conf)
	}
}
