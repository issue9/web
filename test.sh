#!/bin/sh
# Copyright 2018 by caixw, All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

list=$(go list ./...| grep -v /vendor/)

echo '生成 so 文件'
go generate -v $list

echo '执行 go vet'
go vet  -v $list

echo '执行 go test'
go test -v $list
