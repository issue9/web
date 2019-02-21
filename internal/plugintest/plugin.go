// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// +build linux darwin

//go:generate go build -o=./testdata/plugin_1.so -buildmode=plugin ./testdata/plugin1/plugin.go
//go:generate go build -o=./testdata/plugin_2.so -buildmode=plugin ./testdata/plugin2/plugin.go

// Package plugintest 作为插件的功能测试包
//
// NOTE: 该功能如果直接写在 module 包之下，目前版本会报错。
//
// NOTE: 需要注意保持上面的 +build 指令中的值与 pluginOS 变量中的一致。
package plugintest
