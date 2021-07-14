#!/bin/bash

# 转账数据
bin/generate tx --total 100 --split 10 \
--txid bfe1eb7708dc367a4f492574cbc586395effb28d7b642cd882b8420cfff9c66f \
--address dw3RjnTe47G4u6a6hHWCfEhtaDkgdYWTE \
--amount 1000000000000 \
--output ./data/transaction \
--process 1 --concurrency 2 &> log &

# 存证数据
# 单进程 10w tps;
bin/generate evidence --total 1000000 --length 200 \
--output ./data/evidence \
--process 1 --concurrency 20 &> log &
