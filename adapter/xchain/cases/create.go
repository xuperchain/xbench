package cases

import (
	"github.com/xuperchain/xuperbench/adapter/xchain/lib"
	"github.com/xuperchain/xuperbench/common"
)

var (
	Bank = &lib.Acct{}
	Accts = map[int]*lib.Acct{}
	Clis = []*lib.Client{}
	LBank = &lib.LcvAcct{}
	LAccts = map[int]*lib.LcvAcct{}
	LCBank = &lib.LcvContract{}
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
	case common.Relay:
		return &Relay{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	case common.LcvTrans:
		return &LcvTrans{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	case common.LcvInvoke:
		return &LcvInvoke{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	case common.QueryBlock:
		return &QueryBlock{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	case common.QueryTx:
		return &QueryTx{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	case common.QueryAcct:
		return &QueryAcct{
			common.TestCase{
				BCType: common.Xchain,
				Label:  label,
			},
		}
	default:
		panic("unknow testcase of xchain")
	}
}
