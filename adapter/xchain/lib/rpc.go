package lib

import (
	"context"
	"encoding/hex"
	"math/big"
//	"io/ioutil"
	"strconv"
	"time"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
	"github.com/xuperchain/xuperunion/crypto/client"
	"google.golang.org/grpc"
//	"github.com/golang/protobuf/proto"
)

var (
	conn *grpc.ClientConn
	cli pb.XchainClient
)

func Connect(host string) {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithMaxMsgSize(64<<20-1))
	conn, _ = grpc.Dial(host, opts...)
	cli = pb.NewXchainClient(conn)
}

func header() *pb.Header {
	out := &pb.Header{
		Logid: global.Glogid(),
	}
	return out
}

func GetBalance(addr string, bcname string) (*pb.AddressStatus, error) {
	bc := &pb.TokenDetail{
		Bcname: bcname,
	}
	in := &pb.AddressStatus{
		Header: header(),
		Address: addr,
		Bcs: []*pb.TokenDetail{bc},
	}
	out, err := cli.GetBalance(context.Background(), in)
	return out, err
}

func GetFrozenBalance(addr string, bcname string) *pb.AddressStatus {
	bc := &pb.TokenDetail{
		Bcname: bcname,
	}
	in := &pb.AddressStatus{
		Header: header(),
		Address: addr,
		Bcs: []*pb.TokenDetail{bc},
	}
	out, _ := cli.GetFrozenBalance(context.Background(), in)
	return out
}

func GetBlock(blockid string, bcname string) *pb.Block {
	id, _ := hex.DecodeString(blockid)
	in := &pb.BlockID{
		Header: header(),
		Bcname: bcname,
		Blockid: id,
		NeedContent: true,
	}
	out, _ := cli.GetBlock(context.Background(), in)
	return out
}

func QueryTx(txid string, bcname string) *pb.TxStatus {
	tx , _ := hex.DecodeString(txid)
	in := &pb.TxStatus{
		Header: header(),
		Bcname: bcname,
		Txid: tx,
	}
	out, _ := cli.QueryTx(context.Background(), in)
	return out
}

func QueryACL(bcname string, acct string) *pb.AclStatus {
	in := &pb.AclStatus{
		Header: header(),
		Bcname: bcname,
		AccountName: acct,
	}
	out, _ := cli.QueryACL(context.Background(), in)
	return out
}

func GetBlockChains() *pb.BlockChains {
	in := &pb.CommonIn{
		Header: header(),
	}
	out, _ := cli.GetBlockChains(context.Background(), in)
	return out
}

func GetSystemStatus() *pb.SystemsStatusReply {
	in := &pb.CommonIn{
		Header: header(),
	}
	out, _ := cli.GetSystemStatus(context.Background(), in)
	return out
}

func SelectUTXO(f *Acct, bcname string, need string, name string) (*pb.UtxoOutput, error) {
	in := &pb.UtxoInput{
		Header: header(),
		Bcname: bcname,
		Address: name,
		Publickey: f.Pub,
		TotalNeed: need,
		NeedLock: true,
	}
	out, err := cli.SelectUTXO(context.Background(), in)
	if err != nil {
		log.ERROR.Printf("select utxo error %#v", err)
		return nil, err
	}
	return out, nil
}

func FormatTx(from string) *pb.Transaction {
	tx := &pb.Transaction{
		Version: 1,
		Coinbase: false,
		Desc: []byte(""),
		Nonce: global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: from,
	}
	return tx
}

func FormatTxOutput(tx *pb.Transaction, to string, amount string, frozen string) {
	amt, _ := big.NewInt(0).SetString(amount, 10)
	txout := &pb.TxOutput{
		ToAddr: []byte(to),
		Amount: amt.Bytes(),
	}
	if frozen != "0" {
		frz, _ := strconv.ParseInt(frozen, 10, 64)
		txout.FrozenHeight = frz
	}
	tx.TxOutputs = append(tx.TxOutputs, txout)
}

func FormatTxInput(tx *pb.Transaction, bcname string, from *Acct, name string) {
	total := big.NewInt(0)
	for i := range(tx.TxOutputs) {
		amt := big.NewInt(0).SetBytes(tx.TxOutputs[i].GetAmount())
		total.Add(amt, total)
	}
	utxoRes, _ := SelectUTXO(from, bcname, total.String(), name)
	for _, utxo := range utxoRes.UtxoList {
		txInput := &pb.TxInput{
			RefTxid: utxo.RefTxid,
			RefOffset: utxo.RefOffset,
			FromAddr: utxo.ToAddr,
			Amount: utxo.Amount,
		}
		tx.TxInputs = append(tx.TxInputs, txInput)
	}
	utxoTotal, _ := big.NewInt(0).SetString(utxoRes.TotalSelected, 10)
	// fill the charge
	if utxoTotal.Cmp(total) > 0 {
		delta := utxoTotal.Sub(utxoTotal, total)
		txCharge := &pb.TxOutput{
			ToAddr: []byte(name),
			Amount: delta.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txCharge)
	}
}

func FormatTxExt(tx *pb.Transaction, rsp *pb.InvokeResponse, req *pb.InvokeRequest) {
	tx.TxInputsExt = rsp.GetInputs()
	tx.TxOutputsExt = rsp.GetOutputs()
	tx.ContractRequest = req
}

func SignTx(tx *pb.Transaction, from *Acct, name string, bcname string) *pb.TxStatus{
	if name != "" {
		tx.AuthRequire = append(tx.AuthRequire, name + "/" + from.Address)
	} else {
		tx.AuthRequire = append(tx.AuthRequire, from.Address)
	}
	cryptoClient, _ := client.CreateCryptoClient("default")
	signTx, _ := txhash.ProcessSignTx(cryptoClient, tx, []byte(from.Pri))
	signInfo := &pb.SignatureInfo{
		PublicKey: from.Pub,
		Sign: signTx,
	}
	tx.InitiatorSigns = append(tx.InitiatorSigns, signInfo)
	tx.AuthRequireSigns = append(tx.AuthRequireSigns, signInfo)
	tx.Txid, _ = txhash.MakeTransactionID(tx)
	txStatus := &pb.TxStatus{
		Bcname: bcname,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx: tx,
	}
	txStatus.Txid = tx.Txid
	return txStatus
}

func PostTx(txstatus *pb.TxStatus) (*pb.CommonReply, error) {
	out, err := cli.PostTx(context.Background(), txstatus)
	return out, err
}

func PreExec(args map[string][]byte, module string, method string, bcname string,
	contract string) (*pb.InvokeResponse, *pb.InvokeRequest, error) {
	req := &pb.InvokeRequest{
		ModuleName: module,
		MethodName: method,
		Args: args,
	}
	if contract != "" && module != "xkernel" {
		req.ContractName = contract
	}
	in := &pb.InvokeRPCRequest{
		Bcname: bcname,
		Request: req,
		Header: header(),
	}
	out, err := cli.PreExec(context.Background(), in)
	return out.GetResponse(), req, err
}

// to new contract account
//func GenPreExeRes(name string, addr string, chain string) (*pb.InvokeResponse,
//	*pb.InvokeRequest, error) {
//	args := make(map[string][]byte)
//	args["account_name"] = []byte(name)
//	acl := `{
//		"pm": {
//            "rule": 1,
//            "acceptValue": 1.0
//        },
//        "aksWeight": {
//            "` + addr + `": 1.0
//        }
//	}`
//	args["acl"] = []byte(acl)
//	return PreExec(args, "xkernel", "NewAccount", chain, "")
//}
//
//func GenContractExeRes(initargs string, acctname string, contract string, code string,
//	lang string, chain string) (*pb.InvokeResponse, *pb.InvokeRequest, error) {
//	args := make(map[string][]byte)
//	args["account_name"] = []byte(acctname)
//	args["contract_name"] = []byte(contract)
//	desc := &pb.WasmCodeDesc{
//		Runtime: lang,
//	}
//	buf , _ := proto.Marshal(desc)
//	args["contract_desc"] = buf
//	codebuf, _ := ioutil.ReadFile(code)
//	args["contract_code"] = codebuf
//	args["init_args"] = []byte(initargs)
//	return PreExec(args, "xkernel", "Deploy", chain, "")
//}
