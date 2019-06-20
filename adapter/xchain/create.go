package xchain

import (
	"github.com/xuperchain/xuperbench/common"
)

// New return the xchain testcase
func New(label common.CaseType) common.ICaseFace {
	switch label {
	case common.Query:
		return &Query{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	case common.Deal:
		return &Deal{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
//	case common.Open:
//		return &Open{
//			common.TestCase{
//				BCType: common.Xchain,
//				Label:  label,
//			},
//		}
	default:
		panic("unknow testcase of xchain")
	}
}
