package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
)

// 银行账户
var Bank *account.Account

//var Bank = &account.Account{
//	Address: `dw3RjnTe47G4u6a6hHWCfEhtaDkgdYWTE`,
//	PublicKey: `{"Curvname":"P-256","X":71150494877248293798614437171152372361228736891836815976675168211334131079261,"Y":93501855315423594331057555514461624511800705618893328391445695924964114158010}`,
//	PrivateKey: `{"Curvname":"P-256","X":71150494877248293798614437171152372361228736891836815976675168211334131079261,"Y":93501855315423594331057555514461624511800705618893328391445695924964114158010,"D":15507592376131504499689165371014638207897342077694859168158927265802326599966}`,
//}

func init() {
	bank, err := LoadBankAK()
	if err != nil {
		log.Fatalf("load bank keys error: %v", err)
	}

	Bank = bank
}

func LoadBankAK() (*account.Account, error) {
	dir, _ := os.Getwd()
	path := filepath.Join(dir, "./data/bank")
	var addr, err = ioutil.ReadFile(filepath.Join(path, "address"))
	if err != nil {
		return nil, fmt.Errorf("read address error: %v", err)
	}

	priKey, err := ioutil.ReadFile(filepath.Join(path, "private.key"))
	if err != nil {
		return nil, fmt.Errorf("read private.key error: %v", err)
	}

	pubKey, err := ioutil.ReadFile(filepath.Join(path, "public.key"))
	if err != nil {
		return nil, fmt.Errorf("read public.key error: %v", err)
	}

	addInfo := &account.Account{
		Address:    string(addr),
		PublicKey:  string(pubKey),
		PrivateKey: string(priKey),
	}
	return addInfo, nil
}

func LoadAccount(number int) ([]*account.Account, error) {
	if number > 5000 {
		return nil, fmt.Errorf("account not enought: %d", number)
	}
	dir, _ := os.Getwd()
	path := filepath.Join(dir, "./data/account/mnemonic.dat")
	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file error: %s\n", err)
	}
	defer fd.Close()

	var scanner = bufio.NewScanner(fd)
	accounts := make([]*account.Account, number)
	for i := 0; i < number; i++ {
		scanner.Scan()
		ak, err := account.RetrieveAccount(scanner.Text(), 2)
		if err != nil {
			continue
		}

		accounts[i] = ak
	}

	return accounts, nil
}

func GenerateAddress(concurrency int) {
	dir, _ := os.Getwd()
	path := filepath.Join(dir, "./data/account/mnemonic.dat")
	fd, err := os.Open(path)
	if err != nil {
		fmt.Printf("open file error: %s\n", err)
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)

	var buffer bytes.Buffer
	for i := 0; i < concurrency; i++ {
		scanner.Scan()
		ak, err := account.RetrieveAccount(scanner.Text(), 2)
		if err != nil {
			continue
		}

		buffer.WriteString(fmt.Sprintf("%s\n", ak.Address))
	}

	addressPath := filepath.Join(dir, "../data/account/address.dat")
	_ = ioutil.WriteFile(addressPath, buffer.Bytes(), 0644)
}
