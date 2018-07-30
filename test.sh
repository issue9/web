#!/bin/sh
# Copyright 2018 by caixw, All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

echo '生成 so 文件'
go generate ./...

echo '执行 go vet'
go vet ./...

echo '执行 go test'
go test ./...
