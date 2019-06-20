package config

import (
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/adapter/demo"
	"github.com/xuperchain/xuperbench/adapter/xchain"
)

// SetCallBack set the callback while do bench test according to the TestCase info
func SetCallBack(msg []*common.BenchMsg) {
	demoDeal := demo.New(common.Deal)
	demoOpen := demo.New(common.Open)
	xchainDeal := xchain.New(common.Deal)
	xchainQuery := xchain.New(common.Query)
//	xchainOpen := xchain.New(common.Open)

	for _, v := range msg {
		switch v.TestCase {
		case demoDeal.GetTestCase():
			v.CB = demoDeal
		case demoOpen.GetTestCase():
			v.CB = demoOpen
		case xchainDeal.GetTestCase():
			v.CB = xchainDeal
		case xchainQuery.GetTestCase():
			v.CB = xchainQuery
//		case xchainOpen.GetTestCase():
//			v.CB = xchainOpen
		default:
			panic("unknown callback!")
		}
	}
}
