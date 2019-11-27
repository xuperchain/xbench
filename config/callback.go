package config

import (
	"github.com/xuperchain/xuperbench/common"
	"github.com/xuperchain/xuperbench/adapter/demo"
	xchain "github.com/xuperchain/xuperbench/adapter/xchain/cases"
)

// SetCallBack set the callback while do bench test according to the TestCase info
func SetCallBack(msg []*common.BenchMsg) {
	demoDeal := demo.New(common.Deal)
	demoOpen := demo.New(common.Open)
	xchainDeal := xchain.New(common.Deal)
	xchainQuery := xchain.New(common.Query)
	xchainGenerate := xchain.New(common.Generate)
	xchainInvoke := xchain.New(common.Invoke)
	xchainRelay := xchain.New(common.Relay)
	xchainLcvTrans := xchain.New(common.LcvTrans)
	xchainLcvInvoke := xchain.New(common.LcvInvoke)
	xchainQueryBlock := xchain.New(common.QueryBlock)
	xchainQueryTx := xchain.New(common.QueryTx)
	xchainQueryAcct := xchain.New(common.QueryAcct)
	xchainRandCase := xchain.New(common.RandCase)

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
		case xchainGenerate.GetTestCase():
			v.CB = xchainGenerate
		case xchainInvoke.GetTestCase():
			v.CB = xchainInvoke
		case xchainRelay.GetTestCase():
			v.CB = xchainRelay
		case xchainLcvTrans.GetTestCase():
			v.CB = xchainLcvTrans
		case xchainLcvInvoke.GetTestCase():
			v.CB = xchainLcvInvoke
		case xchainQueryBlock.GetTestCase():
			v.CB = xchainQueryBlock
		case xchainQueryTx.GetTestCase():
			v.CB = xchainQueryTx
		case xchainQueryAcct.GetTestCase():
			v.CB = xchainQueryAcct
		case xchainRandCase.GetTestCase():
			v.CB = xchainRandCase
		default:
			panic("unknown callback!")
		}
	}
}
