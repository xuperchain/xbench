package cases

import (
	"encoding/hex"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuperchain/xuperchain/service/pb"
)

func QueryTx(*runner.CallData) ([]*dynamic.Message, error) {
	txid := "ce37f55650348d486babed41b56b90c3a6f660e48f6fec0ee30e54c413bd2e20"
	rawTxid, _ := hex.DecodeString(txid)
	protoMsg := &pb.TxStatus{
		Bcname: "xuper",
		Txid: rawTxid,
	}
	dynamicMsg, err := dynamic.AsDynamicMessage(protoMsg)
	if err != nil {
		return nil, err
	}

	return []*dynamic.Message{dynamicMsg}, nil
}

