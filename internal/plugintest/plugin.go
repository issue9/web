// SPDX-License-Identifier: MIT

// +build linux darwin

//go:generate go build -o=./testdata/plugin_1.so -buildmode=plugin ./testdata/plugin1/plugin.go
//go:generate go build -o=./testdata/plugin_2.so -buildmode=plugin ./testdata/plugin2/plugin.go

// Package plugintest 作为插件的功能测试包
//
// NOTE: 该测试如果直接写在功能所在的包，目前版本会报错。
package plugintest
