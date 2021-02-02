# xbench

Benchmark project for XuperUnion (and other blockchain further)

## Pre-requisites

Make sure following requiremants are satisfied:

* Go (version >= 1.11)
* Network (to get depend modules)

## Build xuperbench

Just `make` it

## Working mode

### Local mode

Send pressure with single endpoint

* Prepare the environment of xuperunion
* Edit the host address of benchmark config
* Put the xbench binary into xuperunion path
* Run `bin/xbench -c conf/gen.conf`

config meanings:

* gen.conf : profiling normal transfer process
* deal.conf : prepare transaction data before, profiling postTx process
* invoke.conf : profiling contract invoke process

### Remote mode

Send pressure distributedly

* Prepare the environment of xuperunion
* Prepare a environment of Redis (accessable for all endpoint)
* Edit the host address of benchmark config
* Edit the broker(redis) address of benchmark config
* Put the xbench binary into xuperunion path (of each endpoint)
* Run `bin/xbench -worker -c conf/remote.conf` on worker endpoints
* Run `bin/xbench -master -c conf/remote.conf` on master endpoint
* You can edit the conf like gen/deal/invoke to test chosen process

## Getting started

### Install xchain

Install xchain, for more detailed instructions, please review the README.md of xuperchain.

 https://github.com/xuperchain/xuperchain

>### Build
>
>Clone the repository
>
>```
>git clone https://github.com/xuperchain/xuperchain
>```
>
>**Note**: `master` branch contains latest features but might be **unstable**. for production use, please checkout our release branch. the latest release branch is `v3.7`.
>
>Enter the xuperchain folder and build the code:
>
>```
>cd xuperchain
>make
>```
>
>Note that if you are using Go 1.11 or later, go modules are used to download 3rd-party dependencies by default. You can also disable go modules and use the prepared dependencies under vendor folder.
>
>Run test:
>
>```
>make test
>```
>
>Use Docker to build xuperchain see [docker build](https://github.com/xuperchain/xuperchain/blob/master/core/scripts/README.md)
>
>### Run
>
>There is an output folder if build successfully. Enter the output folder, create a default chain firstly:
>
>```
>cd ./output
>./xchain-cli createChain
>```
>
>By doing this, a blockchain named "xuper" is created, you can find the data of this blockchain at `./data/blockchain/xuper/`.
>
>Then start the node and run XuperChain full node servers:
>
>```
>nohup ./xchain &
>```
>
>By default, the `xuper` chain will produce a block every 3 seconds, try the following command to see the `trunkHeight` of chain and make sure it's growing.
>
>```
>./xchain-cli status
>```

###  Account and contract

This step is only for contract  test. You should follow this step，if you want to test contract invoke or contract query. If you do not , just ignore it and go to step "Configuration".

Create xbench test  account. 

```
./xchain-cli  account new --account 1123581321345589 -H :17101 --fee 1000
```

Write you own  contract.  Or use xuperchain contract examples, In this document, we use counter.cc for contract invoke testing

```
# contract examples
github.com/xuperchain/xuperchain/tree/master/core/contractsdk/cpp/example/counter.cc
```

Compile contract. 

```
#compile xdev first
go build -o xdev github.com/xuperchain/xuperchain/core/cmd/xdev
#compile cpp counter
export XDEV_ROOT=`pwd`/core/contractsdk/cpp && ./xdev build -o wasm_cpp_counter.wasm  core/contractsdk/cpp/example/counter.cc
```

Deploy smart contract.

```
./xchain-cli wasm deploy --account XC1123581321345589@xuper --cname wasm_cpp -a '{"creator":"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"}' contract/wasm_cpp_counter.wasm --fee 152806  -H 10.117.131.15:17101
```

###  configuration

Xbench use conf  file to define benchmark behaviors. Here is an example of conf 

```json
{
  //tested block chain type(xchain/fabric)
    "type": "xchain",
  //concurrent num of transaction 
    "workNum": 10,
  //working mode (local/remote)
    "mode": "local",
  //chain code name
    "chain": "xuper",
  //Encryption plugins（default/schnorr)
    "crypto": "default",
  //tested block chain ip:port
    "host": "127.0.0.1:37101",
  //test behaviors (deal/generate/invoke/query/...)
    "rounds": [
        {
          //test cases
            "label": "deal",
          //requests of every worker
            "number": [ 20 ]
        }
    ]
}

```

Xbench now supports 10 test cases which is defined by label. Here are the explanation of each case.

**deal**:  Test postTx performance in transfer. Default prepare enough TXs and sign them.  

**generate**: Test account transfer performance.Default create enough accounts, transfer 1 from these accounts to one test account.

**relay**: Test account transfer performance. Default create enough accounts, transfer 1 to themselves using the last txid, which do not need to selectUTXO.

**query**: Test contract query performance. Default deploy `counter`and invoke `increase` method to increase key, query this key to test.

**invoke**: Test contract invoke performance. Default deploy `counter`,invoke `increase` method to test

**querytx**: Test tx query performace. Query the previous reftxid, query the first one if reach the end.

**queryblock**: Test block query performance. Query the previous pre block, query the first one if reach the end.

**queryacct**: Test account balance query performance. Default transfer n to an account, 

**lcvtrans**: Test transfer performnce using sdk. Should deploy endorser. 

**lcvinvoke**: Test contract invoke perfromace using sdk. Also should deploy endorser.

### Run

Test deal performance:

```
bin/xbench -c conf/deal.conf
```

Test contract invoke performance:

```
bin/xbench -c conf/invoke.conf
```

### Report 

After running , you can get a html test report.


##  Development

 update later....

