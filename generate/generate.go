package generate

import "github.com/xuperchain/xuperchain/service/pb"

type Generator interface {
	Generate() chan []*pb.Transaction
}

