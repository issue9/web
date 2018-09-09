// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// +build linux darwin

//go:generate go build -o=./testdata/plugin_1.so -buildmode=plugin ./testdata/plugin1/plugin.go
//go:generate go build -o=./testdata/plugin_2.so -buildmode=plugin ./testdata/plugin2/plugin.go
//go:generate go build -o=./testdata/plugin-3.so -buildmode=plugin ./testdata/plugin3/plugin.go

package modules

// 此文件仅作为生成测试 *.so 文件，
// NOTE: 需要注意保持上面的 +build 指令中的值与 pluginOS 变量中的一致。
