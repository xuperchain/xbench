package cases

import (
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/adapter/fabric/lib"
)

var (
	Clis = []*lib.FabricNode{}
	SDK_CONF = "conf/fabric/sdk.yaml"
	USER = "User1"
	CHAINCODE = "counter"
)

func New(label common.CaseType) common.ICaseFace {
	switch label {
	case common.Query:
		return &Query{
			common.TestCase{
				BCType: common.Fabric,
				Label:  label,
			},
		}
	case common.Deal:
		return &Deal{
			common.TestCase{
				BCType: common.Fabric,
				Label:  label,
			},
		}
	default:
		panic("unknow testcase of fabric")
	}
}
