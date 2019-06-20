package xuperbench

import (
	"testing"
	"time"
	"github.com/xuperchain/xuperbench/benchmark"
	"github.com/xuperchain/xuperbench/config"
	"github.com/xuperchain/xuperbench/log"
)

func TestIntegrationRemoteMode(t *testing.T) {
	conf := config.ParseConfig("test_remote.json")
	go benchmark.Subscribe(conf)
	go benchmark.Subscribe(conf)
	time.Sleep(5 * time.Second)
	go func() {
		benchmark.Publish(conf)
		benchmark.BackendProf(conf)
	}()

	log.INFO.Println("TestIntegrationRemoteMode successful!\n")
}

func TestIntegrationLocalMode(t *testing.T) {
	conf := config.ParseConfig("test_local.json")
	benchmark.BenchRun(conf)

	log.INFO.Println("TestIntegrationLocalMode successful!\n")
}
