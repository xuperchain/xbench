#!/bin/bash

# --rps=100 \
# --load-schedule=step --load-start=500 --load-step=100 --load-step-duration=20s \
# --concurrency-schedule=step --concurrency-start=100 --concurrency-step=50 --concurrency-end=300 --concurrency-step-duration=30s \
# --async \

# counter合约
./bench -n 1000000 -c 10 \
--insecure \
--call=pb.Xchain.PostTx \
--proto=./pb/xchain.proto \
--import-paths=./pb/googleapis \
--tags='{"module_name":"wasm","contract_name":"counter","method_name":"increase"}' \
10.117.130.40:36301

# short_content合约
./bench -n 1000000 -c 100 \
--insecure \
--call=pb.Xchain.PostTx \
--proto=./pb/xchain.proto \
--import-paths=./pb/googleapis \
--tags='{"module_name":"wasm","contract_name":"short_content","method_name":"storeShortContent","length":"256"}' \
--load-schedule=step --load-start=500 --load-step=100 --load-step-duration=20s \
10.117.130.40:36301

xchain-cli wasm query short_content --method queryByTopic -a '{"user_id":"xuperos","topic":"g0c0"}'
