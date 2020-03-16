package lib

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

type FabricNode struct {
	cli *channel.Client
	peer channel.RequestOption
}

func packArgs(paras []string) [][]byte {
	var args [][]byte
	for _, k := range paras {
		args = append(args, []byte(k))
	}
	return args
}

func InitSDK(path string) (*fabsdk.FabricSDK, error) {
	return fabsdk.New(config.FromFile(path))
}

func CreateNode(ch string, user string, peername string, sdk *fabsdk.FabricSDK) (*FabricNode, error) {
	ccp := sdk.ChannelContext(ch, fabsdk.WithUser(user))
	cc, err := channel.New(ccp)
	if err != nil {
		return nil, err
	}
	peer := channel.WithTargetEndpoints(peername)
	node := &FabricNode{
		cli: cc,
		peer: peer,
	}
	return node, nil
}

func (n *FabricNode) Query(cc string, fcn string, params []string) (channel.Response, error) {
	args := packArgs(params)
	req := channel.Request{
		ChaincodeID: cc,
		Fcn: fcn,
		Args: args,
	}
	return n.cli.Query(req, n.peer)
}

func (n *FabricNode) Execute(cc string, fcn string, params []string) (channel.Response, error) {
	args := packArgs(params)
	req := channel.Request{
		ChaincodeID: cc,
		Fcn: fcn,
		Args: args,
	}
	return n.cli.Execute(req, n.peer)
}
