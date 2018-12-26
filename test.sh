#!/bin/sh
# Copyright 2018 by caixw, All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

# 可以传递两个参数：
# -c 表示需要输出代码测试覆盖率到文件 coverage.txt 中；
# -v 表示需要输出详细的执行信息到终端。

list=$(go list ./...)

while getopts 'vc' OPT; do
    case $OPT in
        v)
            v=-v;;
        c)
            c='-coverprofile=coverage.txt -covermode=atomic';;
        ?)
            echo '未知的参数'
    esac
done

echo '生成 so 文件'
GO111MODULE=on go generate $v $list

echo '执行 go test'
GO111MODULE=on go test $v $c $list
