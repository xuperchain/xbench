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
# 配置文件
bin/xbench --config=conf/transfer.yaml

###############################################################################
# 合约
# counter合约
bin/xbench --config=conf/contract/counter.yaml
# short_content合约
bin/xbench --config=conf/contract/short_content.yaml

###############################################################################
# 生成离线交易
# 转账数据
bin/generate tx --total 1000000 \
--host 127.0.0.1:37101 \
--amount 100000000 \
--output ./data/transaction \
--concurrency 20

# 存证数据
# 提交存证数据需要nofee模式下运行
bin/generate evidence --total 1000000 \
--output ./data/evidence \
--length 256 \
concurrency 20

# 合约数据, nofee模式下运行
# counter合约离线交易
bin/generate contract --config=conf/generate/counter.yaml

# 合约数据, nofee模式下运行
# short_content合约离线交易
bin/generate contract --config=conf/generate/counter.yaml

# 使用离线交易发压
bin/xbench --config=conf/file.yaml
bin/xbench -n 500000 -c 100 \
--tags='{
        "benchmark": "file",
        "path": "./data/transaction"
    }' \
127.0.0.1:37101