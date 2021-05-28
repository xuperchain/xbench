#!/bin/bash

# 脚本绝对路径
AbsPath=$(cd $(dirname "$BASH_SOURCE"); pwd)
# 工作路径
WorkRoot=$AbsPath/..
# 编译产出路径
PackagePath=${1:-$WorkRoot/../xuperos/output}
# 节点配置路径
NodePath=$WorkRoot/node
# 部署路径(本地)
DeployPath=$WorkRoot/deploy

alias info='echo INFO $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'
alias error='echo ERROR $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'

function deployNode() {
    node=$1

    # build
    if [ ! -d "$DeployPath/$node" ]; then
        mkdir "$DeployPath/$node"
    fi

    cp -r "$PackagePath/bin" "$DeployPath/$node/bin"
    cp -r "$PackagePath/control.sh" "$DeployPath/$node"
    cp -r "$NodePath/$node/conf" "$DeployPath/$node/conf"
    cp -r "$NodePath/$node/data" "$DeployPath/$node/data"
    cp -r "$NodePath/genesis" "$DeployPath/$node/data/genesis"

    info "deploy local $node done"
}

function deploy() {
    # check compile output package
    if [ ! -d "$PackagePath/bin" ];then
        error "please compile first"
        exit 1
    fi

    # make deploy dir
    rm -rf "${DeployPath:?}/"*
    mkdir -p "$DeployPath"

    # deploy node
    for filename in $NodePath/*; do
        node=${filename##*/}
        [[ "$node" == node* ]] && deployNode "$node"
    done

    info "deploy local done"
}

# 部署本地节点
deploy "$@"

# 部署远程节点
info "deploy node1: 10.117.130.40"
scp -r "$DeployPath/node1" bench@10.117.130.40:/home/bench
info "deploy node2: 10.117.135.39"
scp -r "$DeployPath/node2" bench@10.117.135.39:/home/bench
info "deploy node3: 10.117.131.15"
scp -r "$DeployPath/node3" bench@10.117.131.15:/home/bench

# 部署发压节点
info "deploy bench: 10.117.135.37"
scp -r "$WorkRoot/output" bench@10.117.135.37:/home/bench/bench
scp -r "$DeployPath/node1" bench@10.117.135.37:/home/bench/bench/node

# 部署远程节点
#scp -r "$WorkRoot/../xuperchain/output.tar.gz" bench@10.117.130.40:/home/bench
#scp -r "$WorkRoot/../xuperchain/output.tar.gz" bench@10.117.135.39:/home/bench
#scp -r "$WorkRoot/../xuperchain/output.tar.gz" bench@10.117.131.15:/home/bench

#info "deploy node1: 10.117.130.40"
#scp -r "$WorkRoot/xchain/node1"/* bench@10.117.130.40:/home/bench/xuperchain
#scp -r "$WorkRoot/contract"/* bench@10.117.130.40:/home/bench/xuperchain/contract
#info "deploy node2: 10.117.135.39"
#scp -r "$WorkRoot/xchain/node2"/* bench@10.117.135.39:/home/bench/xuperchain
#info "deploy node3: 10.117.131.15"
#scp -r "$WorkRoot/xchain/node3"/* bench@10.117.131.15:/home/bench/xuperchain
