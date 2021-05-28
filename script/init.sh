#!/bin/bash

#-------------------------------
# 自动测试功能点：
#    1.搭建网络环境
#    2.转账
#    3.合约
#    4.权限
#------------------------------

# 脚本绝对路径
AbsPath=$(cd $(dirname "$BASH_SOURCE"); pwd)
# 根路径
WorkRoot=$AbsPath/..
# 工作路径：所有命令在node路径下运行
WorkPath=$WorkRoot/node
# 合约路径
ContractPath=$WorkRoot/contract
# 节点host
Node1="10.117.130.40:36301"
Node2="10.117.135.39:36302"
Node3="10.117.131.15:36303"

alias xchain-cli='$WorkPath/bin/xchain-cli -H $Node1'
alias info='echo INFO $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'
alias error='echo ERROR $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'

# account
function account() {
  ## 账户
#  xchain-cli account newkeys --output data/alice || exit
#  xchain-cli transfer --to "$(cat data/alice/address)" --amount 10000000 || exit
#  balance=$(xchain-cli account balance --keys data/alice)
#  info "account $(cat data/alice/address) balance $balance"

  ## 合约账户
  xchain-cli account new --account 1111111111111111 --fee 1000 || exit
  xchain-cli transfer --to XC1111111111111111@xuper --amount 100000001 || exit
  balance=$(xchain-cli account balance XC1111111111111111@xuper)
  info "account XC1111111111111111@xuper balance $balance"
}

# contract
function contract() {
  # short_content
  info "contract short_content"
  xchain-cli wasm deploy $ContractPath/short_content.wasm --cname short_content \
            --account XC1111111111111111@xuper \
            --runtime c -a '{"creator": "xuper"}' --fee 158652 || exit
  info "contract short_content invoke"
  xchain-cli wasm invoke short_content --method storeShortContent \
  -a '{"user_id":"user_id","topic":"topic","title":"title","content":"content"}' --fee 120
  xchain-cli wasm query short_content --method queryByUser -a '{"user_id":"user_id"}'
  xchain-cli wasm query short_content --method queryByTopic -a '{"user_id":"user_id","topic":"topic"}'
  xchain-cli wasm query short_content --method queryByTitle -a '{"user_id":"user_id","topic":"topic","title":"title"}'

  # counter
  info "contract counter"
  xchain-cli wasm deploy $ContractPath/counter.wasm --cname counter \
            --account XC1111111111111111@xuper \
            --runtime c -a '{"creator": "xuper"}' --fee 155537 || exit
  info "contract counter invoke"
  xchain-cli wasm invoke counter --method increase -a '{"key":"test"}' --fee 100
  xchain-cli wasm query counter --method get -a '{"key":"test"}'

  # 查询用户部署的合约
  info "contract XC1111111111111111@xuper"
  xchain-cli account contracts --account XC1111111111111111@xuper
  info "contract $(cat data/keys/address)"
  xchain-cli account contracts --address $(cat data/keys/address)
}

# 内置合约
function builtin() {
      # reserved_contracts
  info "contract reserved unified_check"
  xchain-cli wasm deploy $ContractPath/unified_check --cname unified_check \
            --account XC1111111111111111@xuper \
            --runtime c -a '{"creator": "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"}' --fee 164735 || exit
  xchain-cli wasm invoke unified_check --method register_aks \
            -a '{"aks":"SmJG3rH2ZzYQ9ojxhbRCPwFiE9y6pD1Co,iYjtLcW6SVCiousAb5DFKWtWroahhEj4u"}' --fee 155 || exit

  # forbidden_contract
  info "contract forbidden"
  xchain-cli wasm deploy $ContractPath/forbidden --cname forbidden \
            --account XC1111111111111111@xuper \
            --runtime c -a '{"creator": "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"}' --fee 155679 || exit
  sleep 3s
}

function height() {
  height1=$(xchain-cli status -H $Node1 | grep trunkHeight | awk '{print $2}')
  height2=$(xchain-cli status -H $Node2 | grep trunkHeight | awk '{print $2}')
  height3=$(xchain-cli status -H $Node3 | grep trunkHeight | awk '{print $2}')

  info "height1=$height1 height2=$height2 height3=$height3"
  diff=$((2*height1-height2-height3))
  if [ $diff -gt 3 ]; then
		error "height inconsistency: height1=$height1 height2=$height2 height3=$height3" && exit
  fi
}

function init() {
  info "init account"
  account

  info "init contract"
  contract

  info "init height"
  height

  info "init done"
}

cd "$WorkPath" || exit

case X$1 in
    Xaccount)
        account
        ;;
    Xcontract)
        contract
        ;;
    Xacl)
        acl
        ;;
    Xheight)
        height
        ;;
    Xtransfer)
        transfer
        ;;
    X*)
        init "$@"
esac

cd - || exit