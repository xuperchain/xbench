package xchain

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/xuperchain/xuperbench/log"
)

type Acct struct {
	Address string
	Pub string
	Pri string
}

func getFileContent(file string) string {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.ERROR.Printf("read file error: %s", err)
		return ""
	}
	return string(f)
}

func InitBankAcct() *Acct {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	keyPath := filepath.Join(dir, "../data/keys")
	addr := getFileContent(keyPath + "/address")
	pubkey := getFileContent(keyPath + "/public.key")
	scrkey := getFileContent(keyPath + "/private.key")
	acct := &Acct{
		Address: addr,
		Pub: pubkey,
		Pri: scrkey,
	}
	return acct
}

func CreateTestClients(num int, host string) map[int]*Acct {
	accts := map[int]*Acct{}
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	testKeysPath := filepath.Join(dir, "../data/testkeys/")
	if _, err := os.Stat(testKeysPath); os.IsNotExist(err) {
        os.Mkdir(testKeysPath, 0755)
    }
	for i:=0; i<num; i+=1 {
		tpath := filepath.Join(testKeysPath, strconv.Itoa(i))
		args := fmt.Sprintf("account newkeys -o %s", tpath)
		if _, e := os.Stat(tpath); os.IsNotExist(e) {
			RunCliCmd(args, host)
			log.INFO.Printf("create account %d", i)
		}
		acct := &Acct{
			Address: getFileContent(tpath + "/address"),
			Pub: getFileContent(tpath + "/public.key"),
			Pri: getFileContent(tpath + "/private.key"),
		}
		accts[i] = acct
	}
	return accts
}

func RunCliCmd(args string, host string) string {
	var out bytes.Buffer
	f := strings.Fields(args + " -H " + host)
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	cmd := exec.Command(dir + "/xchain-cli", f...)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
	    return ""
	}
	return out.String()
}
