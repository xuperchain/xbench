#!/bin/bash

# 转账数据
bin/generate tx --total 500000 \
--host 10.117.130.40:32101 \
--output ./data/transaction \
--process 10 --concurrency 10 &> log &

# 存证数据
bin/generate evidence --total 1000000 --length 200 \
--output ./data/evidence \
--process 10 --concurrency 10 &> log &
