#!/bin/sh
# Copyright 2017 by caixw, All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

# 指定工作目录
wd=$(dirname $0)/../cmd/web

# 指定编译日期
date=`date -u '+%Y%m%d'`

# 需要修改变量的地址，若为 main，则指接使用 main，而不是全地址
path=github.com/issue9/web/internal/cmd/version

cd ${wd}

echo '开始编译'
go build -ldflags "-X ${path}.buildDate=${date}" -v
