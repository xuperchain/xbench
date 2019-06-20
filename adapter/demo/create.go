package demo

import (
	"github.com/xuperchain/xuperbench/common"
)

// New return the demo testcase
func New(label common.CaseType) common.ICaseFace {
	switch label {
	case common.Open:
		return &Open{
			common.TestCase{
				BCType: common.Demo,
				Label:  label,
			},
		}
	case common.Deal:
		return &Deal{
			common.TestCase{
				BCType: common.Demo,
				Label:  label,
			},
		}
	default:
		panic("unknow testcase of demo")
	}
}
