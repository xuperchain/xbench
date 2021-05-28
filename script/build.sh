#!/bin/bash

# 脚本绝对路径
AbsPath=$(cd $(dirname "$BASH_SOURCE"); pwd)
# 工作路径
WorkRoot=$AbsPath/..
# 产出路径
Output=$WorkRoot/output

function buildBench() {
    rm -rf "${Output:?}/"*
    mkdir -p "$Output"

    go build -o $Output/bench $WorkRoot/main.go
    cp -r $WorkRoot/conf $Output
    cp -r $WorkRoot/data $Output
    cp -r $WorkRoot/contract $Output
    cp -r $WorkRoot/pb $Output
    cp -r $WorkRoot/script $Output
}

buildBench