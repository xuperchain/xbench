package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xuperchain/xuper-sdk-go/account"
	"github.com/xuperchain/xuper-sdk-go/contract"
	"github.com/xuperchain/xuper-sdk-go/contract_account"
	"github.com/xuperchain/xuper-sdk-go/transfer"
)

// define blockchain node and blockchain name
var (
	contractName = "counter"
	node         = "x.x.x.x:8080" // node ip
	bcname       = "xuper"
)

func testAccount() {
	// create an account for the user,
	// strength 1 means that the number of mnemonics is 12
	// language 1 means that mnemonics is Chinese
	acc, err := account.CreateAccount(1, 1)
	if err != nil {
		fmt.Printf("create account error: %v\n", err)
	}
	// print the account, mnemonics
	fmt.Println(acc)
	fmt.Println(acc.Mnemonic)

	// retrieve the account by mnemonics
	acc, err = account.RetrieveAccount(acc.Mnemonic, 1)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
	}
	fmt.Printf("RetrieveAccount: to %v\n", acc)

	// create an account, then encrypt using password and save it to a file
	acc, err = account.CreateAndSaveAccountToFile("./keys", "123", 1, 1)
	if err != nil {
		fmt.Printf("createAndSaveAccountToFile err: %v\n", err)
	}
	fmt.Printf("CreateAndSaveAccountToFile: %v\n", acc)

	// get the account from file, using password decrypt
	acc, err = account.GetAccountFromFile("keys/", "123")
	if err != nil {
		fmt.Printf("getAccountFromFile err: %v\n", err)
	}
	fmt.Printf("getAccountFromFile: %v\n", acc)
	return
}

func testContractAccount() {
	// retrieve the account by mnemonics
	// Notice !!!
	// parameters should be Mnemonics for your account and language
	account, err := account.RetrieveAccount(Mnemonics, language)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
	}

	// define the name of the conrtact account to be created
	// Notice !!!
	// conrtact account is (XC + 16 length digit + @xuper), like "XC1234567890123456@xuper"
	contractAccount := contractAcc

	// initialize a client to operate the contract account
	ca := contractaccount.InitContractAccount(account, node, bcname)

	// create contract account
	txid, err := ca.CreateContractAccount(contractAccount)
	if err != nil {
		log.Printf("CreateContractAccount err: %v", err)
		os.Exit(-1)
	}
	fmt.Println(txid)
	/*
		// the 2nd way to create contract account
		preSelectUTXOResponse, err := ca.PreCreateContractAccount(contractAccount)
		if err != nil {
			log.Printf("PreCreateContractAccount failed, err: %v", err)
			os.Exit(-1)
		}
		txid, err := ca.PostCreateContractAccount(preSelectUTXOResponse)
		if err != nil {
			log.Printf("PostCreateContractAccount failed, err: %v", err)
			os.Exit(-1)
		}
		log.Printf("txid: %v", txid)
	*/
	return
}

func testTransfer() {
	// retrieve the account by mnemonics
	// Notice !!!
	// parameters should be Mnemonics for your account and language
	acc, err := account.RetrieveAccount(Mnemonics, language)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
	}
	fmt.Printf("account: %v\n", acc)

	// initialize a client to operate the transfer transaction
	trans := transfer.InitTrans(acc, node, bcname)

	// transfer destination address, amount, fee and description
	to := "UgdxaYwTzUjkvQnmeB3VgnGFdXfrsiwFv"
	amount := "200"
	fee := "0"
	desc := ""
	// transfer
	txid, err := trans.Transfer(to, amount, fee, desc)
	if err != nil {
		fmt.Printf("Transfer err: %v\n", err)
	}
	fmt.Printf("transfer tx: %v\n", txid)
	return
}

func testDeployWasmContract() {
	// retrieve the account by mnemonics
	// Notice !!!
	// parameters should be Mnemonics for your account and language
	acc, err := account.RetrieveAccount(Mnemonics, language)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
	}
	fmt.Printf("account: %v\n", acc)

	// set contract account, contract will be installed in the contract account
	// Notice !!!
	contractAccount := contractAcc

	// initialize a client to operate the contract
	wasmContract := contract.InitWasmContract(acc, node, bcname, contractName, contractAccount)

	// set init args and contract file
	args := map[string]string{
		"creator": "xchain",
	}
	codePath := "example/contract_code/counter.wasm"

	// deploy wasm contract
	txid, err := wasmContract.DeployWasmContract(args, codePath, "c")
	if err != nil {
		log.Printf("DeployWasmContract err: %v", err)
		os.Exit(-1)
	}
	fmt.Printf("DeployWasmContract txid: %v\n", txid)

	/*
		// the 2nd way to deploy wasm contract, preDeploy and Post
		preSelectUTXOResponse, err := wasmContract.PreDeployWasmContract(args, codePath, "c")
		if err != nil {
			log.Printf("DeployWasmContract GetPreDeployWasmContractRes failed, err: %v", err)
			os.Exit(-1)
		}
		txid, err := wasmContract.PostWasmContract(preSelectUTXOResponse)
		if err != nil {
			log.Printf("DeployWasmContract PostWasmContract failed, err: %v", err)
			os.Exit(-1)
		}
		log.Printf("txid: %v", txid)
	*/
	return
}

func testInvokeWasmContract() {
	// retrieve the account by mnemonics
	// Notice !!!
	// parameters should be Mnemonics for your account and language
	acc, err := account.RetrieveAccount(Mnemonics, language)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
	}
	fmt.Printf("account: %v\n", acc)

	// initialize a client to operate the contract
	contractAccount := ""
	wasmContract := contract.InitWasmContract(acc, node, bcname, contractName, contractAccount)

	// set invoke function method and args
	args := map[string]string{
		"key": "counter",
	}
	methodName := "increase"

	// invoke contract
	txid, err := wasmContract.InvokeWasmContract(methodName, args)
	if err != nil {
		log.Printf("InvokeWasmContract PostWasmContract failed, err: %v", err)
		os.Exit(-1)
	}
	log.Printf("txid: %v", txid)
	/*
		// the 2nd way to invoke wasm contract, preInvoke and Post
		preSelectUTXOResponse, err := wasmContract.PreInvokeWasmContract(methodName, args)
		if err != nil {
			log.Printf("InvokeWasmContract GetPreMethodWasmContractRes failed, err: %v", err)
			os.Exit(-1)
		}
		txid, err := wasmContract.PostWasmContract(preSelectUTXOResponse)
		if err != nil {
			log.Printf("InvokeWasmContract PostWasmContract failed, err: %v", err)
			os.Exit(-1)
		}
		log.Printf("txid: %v", txid)
	*/
	return
}

func testQueryWasmContract() {
	// retrieve the account by mnemonics
	// Notice !!!
	// parameters should be Mnemonics for your account and language
	acc, err := account.RetrieveAccount(Mnemonics, language)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
	}
	fmt.Printf("account: %v\n", acc)

	// initialize a client to operate the contract
	contractAccount := ""
	wasmContract := contract.InitWasmContract(acc, node, bcname, contractName, contractAccount)

	// set query function method and args
	args := map[string]string{
		"key": "counter",
	}
	methodName := "get"

	// query contract
	preExeRPCRes, err := wasmContract.QueryWasmContract(methodName, args)
	if err != nil {
		log.Printf("QueryWasmContract failed, err: %v", err)
		os.Exit(-1)
	}
	gas := preExeRPCRes.GetResponse().GetGasUsed()
	fmt.Printf("gas used: %v\n", gas)
	for _, res := range preExeRPCRes.GetResponse().GetResponse() {
		fmt.Printf("contract response: %s\n", string(res))
	}
	return
}

func testGetBalance() {
	// retrieve the account by mnemonics
	// Notice !!!
	// parameters should be Mnemonics for your account and language
	acc, err := account.RetrieveAccount(Mnemonics, language)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
	}
	fmt.Printf("account: %v\n", acc)

	// initialize a client to operate the transaction
	trans := transfer.InitTrans(acc, node, bcname)

	// get balance of the account
	balance, err := trans.GetBalance()
	log.Printf("balance %v, err %v", balance, err)
	return
}

func testQueryTx() {
	// initialize a client to operate the transaction
	trans := transfer.InitTrans(nil, node, bcname)
	txid := "3a78d06dd39b814af113dbdc15239e675846ec927106d50153665c273f51001e"

	// query tx by txid
	tx, err := trans.QueryTx(txid)
	log.Printf("query tx %v, err %v", tx, err)
	return
}

func main() {
	testAccount()
	testTransfer()
	testDeployWasmContract()
	testInvokeWasmContract()
	testContractAccount()
	testQueryWasmContract()
	testGetBalance()
	testQueryTx()
	return
}
