package generate

import pb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"

type Generator interface {
	Generate() chan []*pb.Transaction
}

