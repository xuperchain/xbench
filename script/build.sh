#!/bin/bash

# 脚本绝对路径
AbsPath=$(cd $(dirname "$BASH_SOURCE"); pwd)
# 工作路径
WorkPath=$AbsPath/..
# 产出路径
Output=$WorkPath/output

function buildBench() {
    rm -rf "${Output:?}/"*
    mkdir -p "$Output/bin"

    go build -o $Output/bin/xbench $WorkPath/cmd/xbench/main.go
    go build -o $Output/bin/generate $WorkPath/cmd/generate/main.go

    cp -r $WorkPath/conf $Output
    cp -r $WorkPath/data $Output
    cp -r $WorkPath/pb $Output
    cp -r $WorkPath/script $Output

    # TODO rm
    cp -r $WorkPath/contract $Output
    ln -s /home/bench/go/src/github.com/xuperchain/xuperchain/output $Output/node
}

buildBench
