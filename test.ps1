# Copyright 2018 by caixw, All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

# 可以传递两个参数：
# -c 表示需要输出代码测试覆盖率到文件 coverage.txt 中；
# -v 表示需要输出详细的执行信息到终端。

param($c,$v)

$list = go list ./... | ? {$_ -notlike '/vendor/*'}

echo '生成 so 文件'
go generate $v $list

echo '执行 go vet'
go vet $v $list

echo '执行 go test'
go test $v $c $list
