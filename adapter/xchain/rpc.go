package xchain

import (
	"context"
	"encoding/hex"
	"math/big"
	"strconv"
	"time"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
	"github.com/xuperchain/xuperunion/crypto/client"
	"google.golang.org/grpc"
)

var (
	conn *grpc.ClientConn
	cli pb.XchainClient
)

func Connect(host string) {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
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

func SelectUTXO(f *Acct, bcname string, need string) (*pb.UtxoOutput, error) {
	in := &pb.UtxoInput{
		Header: header(),
		Bcname: bcname,
		Address: f.Address,
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

func SplitUTXO(f *Acct, bcname string, part int) (string, error) {
	tx := &pb.Transaction{
		Version: 1,
		Coinbase: false,
		Desc: []byte(""),
		Nonce: global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: f.Address,
	}
	amt:= big.NewInt(1)
	txOutput := &pb.TxOutput{
		ToAddr: []byte(f.Address),
		Amount: amt.Bytes(),
		FrozenHeight: 0,
	}
	for i:=0; i<part; i++ {
		tx.TxOutputs = append(tx.TxOutputs, txOutput)
	}

	need := big.NewInt(int64(part))
	utxoRes, err := SelectUTXO(f, bcname, strconv.Itoa(part))
	if err != nil {
		return "", nil
	}
	for _, utxo := range utxoRes.UtxoList {
		txInput := &pb.TxInput{
			RefTxid: utxo.RefTxid,
			RefOffset: utxo.RefOffset,
			FromAddr: utxo.ToAddr,
			Amount: utxo.Amount,
		}
		tx.TxInputs = append(tx.TxInputs, txInput)
	}
	utxoTotal, ok := big.NewInt(0).SetString(utxoRes.TotalSelected, 10)
	if !ok {
		return "", nil
	}
	if utxoTotal.Cmp(need) > 0 {
		delta := utxoTotal.Sub(utxoTotal, need)
		txCharge := &pb.TxOutput{
			ToAddr: []byte(f.Address),
			Amount: delta.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txCharge)
	}

	tx.AuthRequire = []string{f.Address + "/" + f.Address}

	txStatus := &pb.TxStatus{
		Bcname: bcname,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx: tx,
	}

	// sign the Tx
	cryptoClient, _ := client.CreateCryptoClient("default")
	signTx, _ := txhash.ProcessSignTx(cryptoClient, tx, []byte(f.Pri))
	signInfo := &pb.SignatureInfo{
		PublicKey: f.Pub,
		Sign: signTx,
	}
	txStatus.Tx.InitiatorSigns = append(txStatus.Tx.InitiatorSigns, signInfo)
	txStatus.Tx.AuthRequireSigns = append(txStatus.Tx.AuthRequireSigns, signInfo)
	txStatus.Tx.Txid, _ = txhash.MakeTransactionID(txStatus.Tx)
	txStatus.Txid = txStatus.Tx.Txid

	_, err = cli.PostTx(context.Background(), txStatus)
	return hex.EncodeToString(txStatus.Txid), err
}

func GenTx(f *Acct, taddr string, bcname string, amount string, desc string, frozen string) (*pb.TxStatus, error) {
	// build a Tx struct
	tx := &pb.Transaction{
		Version: 1,
		Coinbase: false,
		Desc: []byte(desc),
		Nonce: global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: f.Address,
	}

	// fill the target output
	froz, _ := strconv.ParseInt(frozen, 10, 64)
	amt, _ := big.NewInt(0).SetString(amount, 10)
	txOutput := &pb.TxOutput{
		ToAddr: []byte(taddr),
		Amount: amt.Bytes(),
		FrozenHeight: froz,
	}
	tx.TxOutputs = append(tx.TxOutputs, txOutput)

	// select the token of fromer
	utxoRes, err := SelectUTXO(f, bcname, amount)
	if err != nil {
		return nil, err
	}

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
	if utxoTotal.Cmp(amt) > 0 {
		delta := utxoTotal.Sub(utxoTotal, amt)
		txCharge := &pb.TxOutput{
			ToAddr: []byte(f.Address),
			Amount: delta.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txCharge)
	}

	tx.AuthRequire = []string{f.Address + "/" + f.Address}

	txStatus := &pb.TxStatus{
		Bcname: bcname,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx: tx,
	}

	// sign the Tx
	cryptoClient, err := client.CreateCryptoClient("default")
	if err != nil {
		return nil, err
	}
	signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(f.Pri))
	if err != nil {
		return nil, err
	}
	signInfo := &pb.SignatureInfo{
		PublicKey: f.Pub,
		Sign: signTx,
	}
	txStatus.Tx.InitiatorSigns = append(txStatus.Tx.InitiatorSigns, signInfo)
	txStatus.Tx.AuthRequireSigns = append(txStatus.Tx.AuthRequireSigns, signInfo)
	txStatus.Tx.Txid, _ = txhash.MakeTransactionID(txStatus.Tx)
	txStatus.Txid = txStatus.Tx.Txid

	return txStatus, nil
}

func PostTx(txstatus *pb.TxStatus) (*pb.CommonReply, error) {
	out, err := cli.PostTx(context.Background(), txstatus)
	return out, err
}
