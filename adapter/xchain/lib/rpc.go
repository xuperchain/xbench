package lib

import (
	"context"
	"encoding/hex"
	"strconv"
	"github.com/xuperchain/xuperbench/log"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
	"github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"google.golang.org/grpc"
)

var (
	conn *grpc.ClientConn
	cli []pb.XchainClient
	cryptotype string
)

type Client struct {
	Conn pb.XchainClient
	BC string
}

func SetCrypto(t string) {
	cryptotype = t
}

func Conn(host string, chain string) *Client {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithMaxMsgSize(64<<20-1))
	c, err := grpc.Dial(host, opts...)
	if err != nil {
		log.ERROR.Printf("connect host error: %s", err)
		return nil
	}
	cli := &Client{
		Conn: pb.NewXchainClient(c),
		BC: chain,
	}
	return cli
}

func header() *pb.Header {
	out := &pb.Header{
		Logid: global.Glogid(),
	}
	return out
}

func nonce() string {
	return global.GenNonce()
}

func (cli *Client) GetBalance(addr string) (*pb.AddressStatus, error) {
	bc := &pb.TokenDetail{
		Bcname: cli.BC,
	}
	in := &pb.AddressStatus{
		Header: header(),
		Address: addr,
		Bcs: []*pb.TokenDetail{bc},
	}
	return cli.Conn.GetBalance(context.Background(), in)
}

func (cli *Client) GetFrozenBalance(addr string) (*pb.AddressStatus, error) {
	bc := &pb.TokenDetail{
		Bcname: cli.BC,
	}
	in := &pb.AddressStatus{
		Header: header(),
		Address: addr,
		Bcs: []*pb.TokenDetail{bc},
	}
	return cli.Conn.GetFrozenBalance(context.Background(), in)
}

func (cli *Client) GetBlock(blockid string) (*pb.Block, error) {
	id, err := hex.DecodeString(blockid)
	if err != nil {
		log.ERROR.Printf("got invalid blockid")
		return nil, err
	}
	in := &pb.BlockID{
		Header: header(),
		Bcname: cli.BC,
		Blockid: id,
		NeedContent: true,
	}
	return cli.Conn.GetBlock(context.Background(), in)
}

func (cli *Client) QueryTx(txid string) (*pb.TxStatus, error) {
	tx, err := hex.DecodeString(txid)
	if err != nil {
		log.ERROR.Printf("got invalid txid")
		return nil, err
	}
	in := &pb.TxStatus{
		Header: header(),
		Bcname: cli.BC,
		Txid: tx,
	}
	return cli.Conn.QueryTx(context.Background(), in)
}

func (cli *Client) QueryACL(acct string) (*pb.AclStatus, error) {
	in := &pb.AclStatus{
		Header: header(),
		Bcname: cli.BC,
		AccountName: acct,
	}
	return cli.Conn.QueryACL(context.Background(), in)
}

func (cli *Client) GetBlockChains() (*pb.BlockChains, error) {
	in := &pb.CommonIn{
		Header: header(),
	}
	return cli.Conn.GetBlockChains(context.Background(), in)
}

func (cli *Client) GetSystemStatus() (*pb.SystemsStatusReply, error) {
	in := &pb.CommonIn{
		Header: header(),
	}
	return cli.Conn.GetSystemStatus(context.Background(), in)
}

func (cli *Client) SelectUTXO(f *Acct, need string, name string) (*pb.UtxoOutput, error) {
	in := &pb.UtxoInput{
		Header: header(),
		Bcname: cli.BC,
		Address: name,
		Publickey: f.Pub,
		TotalNeed: need,
		NeedLock: true,
	}
	return cli.Conn.SelectUTXO(context.Background(), in)
}

func (cli *Client) SignTx(tx *pb.Transaction, from *Acct, account string) *pb.TxStatus {
	if account != "" {
		tx.AuthRequire = append(tx.AuthRequire, account + "/" + from.Address)
	} else {
		tx.AuthRequire = append(tx.AuthRequire, from.Address)
	}
	cryptoClient, _ := client.CreateCryptoClient(cryptotype)
	signTx, _ := txhash.ProcessSignTx(cryptoClient, tx, []byte(from.Pri))
	signInfo := &pb.SignatureInfo{
		PublicKey: from.Pub,
		Sign: signTx,
	}
	tx.InitiatorSigns = append(tx.InitiatorSigns, signInfo)
	tx.AuthRequireSigns = append(tx.AuthRequireSigns, signInfo)
	tx.Txid, _ = txhash.MakeTransactionID(tx)
	return &pb.TxStatus{
		Bcname: cli.BC,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx: tx,
		Txid: tx.Txid,
	}
}

func (cli *Client) PostTx(txstatus *pb.TxStatus) (*pb.CommonReply, error) {
	return cli.Conn.PostTx(context.Background(), txstatus)
}

func (cli *Client) PreExecWithSelectUTXO(acct *Acct, need int64) (*pb.PreExecWithSelectUTXOResponse, error) {
	content := hash.DoubleSha256([]byte(cli.BC + acct.Address + strconv.FormatInt(need, 10) + "true"))
	cryptoClient, err := client.CreateCryptoClient(cryptotype)
	if err != nil {
		log.ERROR.Printf("create cryptoclient error")
		return nil, err
	}
	pri, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(acct.Pri))
	sign, _ := cryptoClient.SignECDSA(pri, content)
	signInfo := &pb.SignatureInfo{
		PublicKey: acct.Pub,
		Sign: sign,
	}
	authrequires := []string{acct.Address}
	req := &pb.InvokeRPCRequest{
		Header: header(),
		Bcname: cli.BC,
		Requests: []*pb.InvokeRequest{},
		Initiator: acct.Address,
		AuthRequire: authrequires,
	}
	in := &pb.PreExecWithSelectUTXORequest{
		Header: header(),
		Bcname: cli.BC,
		Address: acct.Address,
		TotalAmount: need,
		SignInfo: signInfo,
		NeedLock: true,
		Request: req,
	}
	return cli.Conn.PreExecWithSelectUTXO(context.Background(), in)
}

func (cli *Client) PreExecWithSelectUTXOContract(acct *Acct, args map[string][]byte, module string, method string, contract string) (*pb.PreExecWithSelectUTXOResponse, error) {
	fromaddr := ""
	irq := &pb.InvokeRequest{
		ModuleName: module,
		MethodName: method,
		Args: args,
	}
	if contract != "" && module != "xkernel" {
		irq.ContractName = contract
	}
	irqs := []*pb.InvokeRequest{irq}
	authrequires := []string{}
	acctname, ok := args["account_name"]
	if ok {
		authrequires = append(authrequires, string(acctname) + "/" + acct.Address)
	} else {
		authrequires = append(authrequires, acct.Address)
	}
	if module == "xkernel" && method == "Deploy" {
		fromaddr = string(acctname)
	} else {
		fromaddr = acct.Address
	}
	req := &pb.InvokeRPCRequest{
		Header: header(),
		Bcname: cli.BC,
		Requests: irqs,
		Initiator: acct.Address,
		AuthRequire: authrequires,
	}
	content := hash.DoubleSha256([]byte(cli.BC + fromaddr + "0" + "true"))
	cryptoClient, err := client.CreateCryptoClient(cryptotype)
	if err != nil {
		log.ERROR.Printf("create cryptoclient error")
		return nil, err
	}
	pri, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(acct.Pri))
	sign, _ := cryptoClient.SignECDSA(pri, content)
	signInfo := &pb.SignatureInfo{
		PublicKey: acct.Pub,
		Sign: sign,
	}
	in := &pb.PreExecWithSelectUTXORequest{
		Header: header(),
		Bcname: cli.BC,
		Address: fromaddr,
		TotalAmount: 0,
		SignInfo: signInfo,
		NeedLock: true,
		Request: req,
	}
	return cli.Conn.PreExecWithSelectUTXO(context.Background(), in)
}

func (cli *Client) PreExec(args map[string][]byte, module string, method string, contract string, addr string) (*pb.InvokeResponse, []*pb.InvokeRequest, error) {
	req := &pb.InvokeRequest{
		ModuleName: module,
		MethodName: method,
		Args: args,
	}
	reqs := []*pb.InvokeRequest{}
	reqs = append(reqs, req)
	if contract != "" && module != "xkernel" {
		req.ContractName = contract
	}
	in := &pb.InvokeRPCRequest{
		Bcname: cli.BC,
		Requests: reqs,
		Header: header(),
	}
	in.Initiator = addr
	acctname, ok := args["account_name"]
	if ok {
		in.AuthRequire = []string{string(acctname) + "/" + addr}
	} else {
		in.AuthRequire = []string{addr}
	}
	out, err := cli.Conn.PreExec(context.Background(), in)
	return out.GetResponse(), out.GetResponse().GetRequests(), err
}


