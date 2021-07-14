module github.com/xuperchain/xbench

go 1.16

require (
	github.com/alecthomas/kingpin v1.3.8-0.20191105203113-8c96d1c22481
	github.com/bojand/ghz v0.94.0
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jhump/protoreflect v1.8.2
	github.com/spf13/cobra v1.0.0
	github.com/xuperchain/xuper-sdk-go/v2 v2.0.0-20210712115836-a7b41261e93a
	github.com/xuperchain/xuperchain v0.0.0-20210708031936-951e4ade7bdd
	github.com/xuperchain/xupercore v0.0.0-20210608021245-b15f81dd9ecf
	go.uber.org/zap v1.15.0
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.34.0
)

replace github.com/xuperchain/xuperchain => ../xuperchain

replace github.com/xuperchain/xupercore => ../xupercore

replace github.com/xuperchain/xuper-sdk-go/v2 => ../xuper-sdk-go
