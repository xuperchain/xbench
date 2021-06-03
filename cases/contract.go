package cases

import (
	"fmt"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xuperos/common/xupospb/pb"
	"google.golang.org/grpc"
	"strconv"
)

var xchain pb.XchainClient
func Conn(host string) pb.XchainClient {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithMaxMsgSize(64<<20-1))
	c, err := grpc.Dial(host, opts...)
	if err != nil {
		return nil
	}

	return pb.NewXchainClient(c)
}

type ContractArgsFunc func(run *runner.CallData, config runner.Config) map[string][]byte
var ContractArgs = map[string]ContractArgsFunc{
	"storeShortContent": StoreShortContentArgs,
	"increase": IncreaseArgs,
}

//
// user_id: string, 用户名
// topic: string 类别(不超过36个字符)
// title: string, 标题(不超过100个字符)
// content: 具体内容(不超过3000个字符)
//
func StoreShortContentArgs(run *runner.CallData, config runner.Config) map[string][]byte {
	var length int
	if lengthStr, ok := config.Tags["length"]; ok {
		n, _ := strconv.ParseUint(lengthStr, 10, 64)
		length = int(n)
	}

	if length <= 0 || length > 3000 {
		length = 64
	}

	args := map[string][]byte{
		"user_id": []byte(`xuperos`),
		"topic": []byte(run.WorkerID),
		"title": []byte(fmt.Sprintf("title_%d_%s", length, RandBytes(16))),
		"content": RandBytes(length),
	}

	return args
}

func IncreaseArgs(run *runner.CallData, config runner.Config) map[string][]byte {
	args := map[string][]byte{
		//"key": []byte(fmt.Sprintf("test_%s_%s", run.WorkerID, RandBytes(16))),
		"key": []byte(fmt.Sprintf("test_%s", run.WorkerID)),
	}
	return args
}

func MakeDataProvider(config runner.Config) (runner.DataProviderFunc, error) {
	if config.Tags == nil {
		return nil, fmt.Errorf("param nil")
	}

	moduleName, ok := config.Tags["module_name"]
	if !ok {
		return nil, fmt.Errorf("module_name not exist")
	}
	contractName, ok := config.Tags["contract_name"]
	if !ok {
		return nil, fmt.Errorf("contract_name not exist")
	}
	methodName, ok := config.Tags["method_name"]
	if !ok {
		return nil, fmt.Errorf("method_name not exist")
	}

	xchain = Conn(config.Host)
	InitAccount(int(config.C))
	return func(run *runner.CallData) ([]*dynamic.Message, error) {
		args := ContractArgs[methodName]
		if args == nil {
			return nil, fmt.Errorf("contract args not register: %s", methodName)
		}
		
		request := &pb.InvokeRequest{
			ModuleName: moduleName,
			ContractName: contractName,
			MethodName: methodName,
			Args: args(run, config),
		}
		ak := AKs[run.RequestNumber%int64(config.C)]
		tx, err := InvokeContract(request, ak)
		if err != nil {
			return nil, err
		}

		protoMsg := &pb.TxStatus{
			Bcname: "xuper",
			Status: pb.TransactionStatus_UNCONFIRM,
			Tx: tx,
			Txid: tx.Txid,
		}
		dynamicMsg, err := dynamic.AsDynamicMessage(protoMsg)
		if err != nil {
			return nil, err
		}

		return []*dynamic.Message{dynamicMsg}, nil
	}, nil
}
