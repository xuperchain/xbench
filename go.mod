module github.com/xuperchain/xuperbench

require (
	github.com/geoffreybauduin/termtables v0.0.0-20190814081757-aef65af557dc
	github.com/golang/protobuf v1.3.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/struCoder/pidusage v0.1.2
	github.com/xuperchain/xuper-sdk-go v0.0.0-20191114073944-eb3c9f0a67b4
	github.com/xuperchain/xuperunion v0.0.0-20191120083657-e6ec20b91828
	google.golang.org/grpc v1.24.0
	github.com/hyperledger/burrow v0.30.5
)

replace (
	golang.org/x/net => github.com/golang/net v0.0.0-20191119073136-fc4aabc6c914
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20191115221424-83cc0476cb11
	google.golang.org/grpc => github.com/grpc/grpc-go v1.24.0
	github.com/hyperledger/burrow => github.com/xuperchain/burrow v0.30.6-0.20200922024403-90193b5a35dd
)
