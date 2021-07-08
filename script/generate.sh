#!/bin/bash

# 转账数据
bin/generate tx --total 1000000 --split 100 \
--txid 003b1dcc3a6d32197d3b08c95a24b77b43426835356fa2e2498f3bd6420f3aee \
--address dw3RjnTe47G4u6a6hHWCfEhtaDkgdYWTE \
--amount 1000000000000 \
--output ./data/transaction \
--process 10 --concurrency 20 &> log &

# 存证数据
# 单进程 10w tps;
bin/generate evidence --total 1000000 --length 200 \
--output ./data/evidence \
--process 1 --concurrency 20 &> log &
