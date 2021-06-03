package cases

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/account"
	"github.com/xuperchain/xuperos/common/xupospb/pb"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var BankAK = &account.Account{
	Address: `dw3RjnTe47G4u6a6hHWCfEhtaDkgdYWTE`,
	PublicKey: `{"Curvname":"P-256","X":71150494877248293798614437171152372361228736891836815976675168211334131079261,"Y":93501855315423594331057555514461624511800705618893328391445695924964114158010}`,
	PrivateKey: `{"Curvname":"P-256","X":71150494877248293798614437171152372361228736891836815976675168211334131079261,"Y":93501855315423594331057555514461624511800705618893328391445695924964114158010,"D":15507592376131504499689165371014638207897342077694859168158927265802326599966}`,
}

var AKs []*account.Account

func InitAccount(concurrency int) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	path := filepath.Join(dir, "data/account/mnemonic.dat")
	fd, err := os.Open(path)
	if err != nil {
		fmt.Printf("open file error: %s\n", err)
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)

	AKs = make([]*account.Account, concurrency)
	for i:=0; i<concurrency; i++ {
		scanner.Scan()
		ak, err := account.RetrieveAccount(scanner.Text(), 2)
		if err != nil {
			continue
		}

		AKs[i] = ak
		tx, err := TransferWithSplit(BankAK, ak.Address, "100000000", int64(concurrency*10))
		in := &pb.TxStatus{
			Bcname: "xuper",
			Status: pb.TransactionStatus_UNCONFIRM,
			Tx: tx,
			Txid: tx.Txid,
		}
		//PrintTx(tx)
		_, err = xchain.PostTx(context.Background(), in)
		if err != nil {
			panic(err)
		}
	}
	log.Print("InitAccount done")
}

func GenAddress(concurrency int) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	path := filepath.Join(dir, "data/account/mnemonic.dat")
	fd, err := os.Open(path)
	if err != nil {
		fmt.Printf("open file error: %s\n", err)
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)

	var buffer bytes.Buffer
	AKs = make([]*account.Account, concurrency)
	for i:=0; i<concurrency; i++ {
		scanner.Scan()
		ak, err := account.RetrieveAccount(scanner.Text(), 2)
		if err != nil {
			continue
		}

		AKs[i] = ak
		buffer.WriteString(fmt.Sprintf("%s\n", ak.Address))
	}

	addressPath := filepath.Join(dir, "data/account/address.dat")
	_ = ioutil.WriteFile(addressPath, buffer.Bytes(), 0644)
}