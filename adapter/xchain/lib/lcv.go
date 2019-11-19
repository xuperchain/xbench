package lib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuper-sdk-go/account"
	"github.com/xuperchain/xuper-sdk-go/config"
	"github.com/xuperchain/xuper-sdk-go/contract"
	"github.com/xuperchain/xuper-sdk-go/contract_account"
	"github.com/xuperchain/xuper-sdk-go/transfer"
)

type LcvAcct struct {
	Acct *account.Account
	Tran *transfer.Trans
}

type LcvContract struct {
	CAcct *contractaccount.ContractAccount
	Contract *contract.WasmContract
}

func GenLcvConfig(endorse, xcheck, bank string) *config.CommConfig {
	compilanceCfg := &config.ComplianceCheckConfig{
		ComplianceCheckEndorseServiceFee: 100,
		ComplianceCheckEndorseServiceFeeAddr: bank,
		ComplianceCheckEndorseServiceAddr: xcheck,
	}
	cfg := &config.CommConfig{
		EndorseServiceHost: endorse,
		ComplianceCheck: *compilanceCfg,
	}
	return cfg
}

func CreateLcvAcct(bcname string, host string) *LcvAcct {
	acct, _ := account.CreateAccount(1, 2)
	trans := transfer.InitTrans(acct, host, bcname)
	return &LcvAcct{
		Acct: acct,
		Tran: trans,
	}
}

func RetrieveAccount(mnemonic string, bcname string, host string) *LcvAcct {
	acct, _ := account.RetrieveAccount(mnemonic, 2)
	trans := transfer.InitTrans(acct, host, bcname)
	return &LcvAcct{
		Acct: acct,
		Tran: trans,
	}
}

func BatchRetrieve(accts map[int]*LcvAcct, size int, bcname string, host string) {
	if size > 5000 {
		return
	}
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	path := filepath.Join(dir, "../data/mnemonic.dat")
	fd, _ := os.Open(path)
	defer fd.Close()
	scanner := bufio.NewScanner(fd)
	for i:=0; i<size; i++ {
		scanner.Scan()
		accts[i] = RetrieveAccount(scanner.Text(), bcname, host)
	}
}

func GetAccountFromFile(path, bcname, host string) *LcvAcct {
	if path == "" {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		path = filepath.Join(dir, "../data/keys")
	}
	adr := getFileContent(path + "/address")
	pub := getFileContent(path + "/public.key")
	pri := getFileContent(path + "/private.key")
	acct := &account.Account{
		Address:    adr,
		PublicKey:  pub,
		PrivateKey: pri,
	}
	trans := transfer.InitTrans(acct, host, bcname)
	return &LcvAcct{
		Acct: acct,
		Tran: trans,
	}
}

func (l *LcvAcct) Transfer(to, amount, fee, desc string) (string, error) {
	return l.Tran.Transfer(to, amount, fee, desc)
}

func (l *LcvAcct) SetCfg(cfg *config.CommConfig) {
	l.Tran.Xchain.Cfg = cfg
}

func CreateLcvContract(acct *LcvAcct, accountname, contractname string) *LcvContract {
	host := acct.Tran.Xchain.XchainSer
	bcname := acct.Tran.Xchain.ChainName
	ca := contractaccount.InitContractAccount(acct.Acct, host, bcname)
	ca.CreateContractAccount(accountname)
	aname := fmt.Sprintf("XC%s@%s", accountname, bcname)
	cract := contract.InitWasmContract(acct.Acct, host, bcname, contractname, aname)
	return &LcvContract{
		CAcct: ca,
		Contract: cract,
	}
}

func (l *LcvContract) SetCfg(cfg *config.CommConfig) {
	l.CAcct.Xchain.Cfg = cfg
	l.Contract.Xchain.Cfg = cfg
}

func (l *LcvContract) SetContractName(name string) {
	l.Contract.ContractName = name
}

func (l *LcvContract) Deploy(args map[string]string, codepath string, runtime string) (string, error) {
	return l.Contract.DeployWasmContract(args, codepath, runtime)
}

func (l *LcvContract) Invoke(methodName string, args map[string]string) (string, error) {
	return l.Contract.InvokeWasmContract(methodName, args)
}

func (l *LcvContract) Query(methodName string, args map[string]string) (*pb.InvokeRPCResponse, error) {
	return l.Contract.QueryWasmContract(methodName, args)
}

func LcvInitIdentity(addrs []string, l *LcvContract) (string, error) {
	args := make(map[string]string)
	args["aks"] = strings.Join(addrs, ",")
	l.SetContractName("unified_check")
	return l.Invoke("register_aks", args)
}
