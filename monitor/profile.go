package monitor

import (
	"fmt"
	"github.com/struCoder/pidusage"
)

type record struct {
	max   float64
	min   float64
	total float64
}

type StatInfo struct {
	cpu   *record
	mem   *record
	count int
}

func (s StatInfo) String() string {
	return fmt.Sprintf("total: %v, cpu: [max: %v, min: %v], mem: [max: %v, min: %v]",
		s.count, s.cpu.max, s.cpu.min, s.mem.max, s.mem.min)
}

func NewInfo() *StatInfo {
	v := &StatInfo{}
	v.cpu = new(record)
	v.mem = new(record)

	return v
}

func (s *StatInfo) Add(x *pidusage.SysInfo) {
	if s.cpu.max < x.CPU {
		s.cpu.max = x.CPU
	} else if s.cpu.min > x.CPU {
		s.cpu.min = x.CPU
	}
	s.cpu.total += x.CPU

	if s.mem.max < x.Memory {
		s.mem.max = x.Memory
	} else if s.mem.min > x.Memory {
		s.mem.min = x.Memory
	}
	s.mem.total += x.Memory

	s.count++
}
