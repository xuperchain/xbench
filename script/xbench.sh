#!/bin/bash

###############################################################################
# 发压参数
# 恒定rps发压
# --rps=100 \
# rps梯度发压
# --load-schedule=step --load-start=500 --load-step=100 --load-step-duration=20s \
# 并发梯度发压
# --concurrency-schedule=step --concurrency-start=100 --concurrency-step=50 --concurrency-end=300 --concurrency-step-duration=30s \
# 异步发压
# --async \

###############################################################################
# 转账
## 配置文件
bin/xbench --config=conf/transfer.json
## 命令行参数
bin/xbench -n 500000 -c 100 \
--rps=100 \
--tags='{
        "benchmark": "transfer",
        "amount": "100000000"
    }' \
127.0.0.1:32101

###############################################################################
# 合约
bin/xbench --config=conf/contract/counter.json
bin/xbench -n 500000 -c 10 \
--tags='{
        "benchmark": "contract",
        "amount": "100000000",

        "contract_account": "XC1111111111111111@xuper",
        "code_path": "./contract/counter.wasm",

        "module_name":"wasm",
        "contract_name":"counter",
        "method_name":"increase"
    }' \
127.0.0.1:32101

###############################################################################
# 生成离线交易
# 转账数据
bin/generate tx --total 1000000 \
--host 127.0.0.1:32101 \
--amount 100000000 \
--output ./data/transaction \
--process 10 --concurrency 10

# 存证数据
# 提交存证数据需要nofee模式下运行
bin/generate evidence --total 1000000 \
--output ./data/evidence \
--length 256 \
--process 10 --concurrency 10

# 使用离线交易发压
bin/xbench --config=conf/file.json
bin/xbench -n 500000 -c 100 \
--tags='{
        "benchmark": "file",
        "path": "./data/transaction"
    }' \
127.0.0.1:32101
