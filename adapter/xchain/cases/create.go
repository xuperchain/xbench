package cases

import (
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
)

var (
	Bank = &lib.Acct{}
	Accts = map[int]*lib.Acct{}
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
	case common.Generate:
		return &Generate{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	case common.Invoke:
		return &Invoke{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	default:
		panic("unknow testcase of xchain")
	}
}
