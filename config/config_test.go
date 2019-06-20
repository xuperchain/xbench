package config

import (
	"reflect"
	"testing"
	"github.com/xuperchain/xuperbench/common"
)

var (
	testLocalConfig = &Config{
		BCType:    common.Demo,
		WorkerNum: 5,
		Mode:      common.LocalMode,
		Rounds: []struct {
			Label    common.CaseType `json:"label"`
			Number   []int         `json:"number,omitempty"`
			Duration []int         `json:"duration,omitempty"`
		}{
			{
				Label:  common.Open,
				Number: []int{100, 200, 150},
			},
			{
				Label:    common.Deal,
				Number:   []int{4000, 5000},
				Duration: []int{5, 6},
			},
		},
	}

	TestRemoteConfig = &Config{
		BCType:        common.Demo,
		WorkerNum:     5,
		Mode:          common.RemoteMode,
		Broker:        "106.12.120.169:6379",
		ResultBackend: "BackendChan",
		PubSubChan:    "PubSubChan",
		Rounds: []struct {
			Label    common.CaseType `json:"label"`
			Number   []int         `json:"number,omitempty"`
			Duration []int         `json:"duration,omitempty"`
		}{
			{
				Label:  common.Open,
				Number: []int{100, 200, 150},
			},
			{
				Label:    common.Deal,
				Number:   []int{4000, 5000},
				Duration: []int{5, 6},
			},
		},
	}

	testErrorConfig = &Config{
		BCType:        common.Demo,
		WorkerNum:     5,
		Mode:          common.RemoteMode,
		Broker:        "106.12.120.169:6379",
		ResultBackend: "BackendChan",
		PubSubChan:    "PubSubChan",
		Rounds: []struct {
			Label    common.CaseType `json:"label"`
			Number   []int         `json:"number,omitempty"`
			Duration []int         `json:"duration,omitempty"`
		}{
			{
				Label:  common.Open,
				Number: []int{100, 200, 150},
			},
			{
				Label:    common.Deal,
				Number:   []int{4000, 5000},
				Duration: []int{5, 6},
			},
		},
	}
)

func TestParseConfig(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		{
			name: "TestLocal",
			args: args{"../test/test_local.json"},
			want: testLocalConfig,
		},
		{
			name: "TestRemote",
			args: args{"../test/test_remote.json"},
			want: TestRemoteConfig,
		},
		{
			name: "TestNotFound",
			args: args{"notfound.json"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseConfig(tt.args.fileName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBenchMsgFromConf(t *testing.T) {
	type args struct {
		conf *Config
	}
	tests := []struct {
		name string
		args args
		want []*common.BenchMsg
	}{
		// TODO: Add test cases.
		{
			name: "TestZeroValue",
			args: args{&Config{}},
			want: []*common.BenchMsg{},
		},
		{
			name: "TestNil",
			args: args{nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBenchMsgFromConf(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBenchMsgFromConf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBenchMsgFromConfigFile(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want []*common.BenchMsg
	}{
		// TODO: Add test cases.
		{
			name: "TestEmpty",
			args: args{""},
			want: nil,
		},
		{
			name: "TestNotFound",
			args: args{"NotFound.json"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBenchMsgFromConfigFile(tt.args.fileName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBenchMsgFromConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
