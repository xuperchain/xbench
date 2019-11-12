package common

type BlockChain string
type CaseType string
type TestMode string
type MsgType string

const (
	Demo   BlockChain = "demo"
	Xchain BlockChain = "xchain"
	EOS    BlockChain = "eos"
	Fabric BlockChain = "fabric"
)

const (
	Open  CaseType = "open"
	Query CaseType = "query"
	Deal  CaseType = "deal"
	Generate CaseType = "generate"
	Invoke CaseType = "invoke"
	Relay CaseType = "relay"
)

const (
	LocalMode  TestMode = "local"
	RemoteMode TestMode = "remote"
	FunctionMode TestMode = "function"
)

const (
	MonMsg MsgType = "MonMsg"
	TpsMsg MsgType = "TpsMsg"
)
