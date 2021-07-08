#!/bin/bash

# 转账数据
bin/generate tx --total 50000000 --split 1000 --process 5 \
--txid 003b1dcc3a6d32197d3b08c95a24b77b43426835356fa2e2498f3bd6420f3aee \
--address dw3RjnTe47G4u6a6hHWCfEhtaDkgdYWTE \
--amount 1000000000000 &> log &

# 存证数据
# 单进程 10w tps;
bin/generate evidence --total 10000000 --length 200 \
--output ./data/evidence \
--process 1 --concurrency 20 &> log &
