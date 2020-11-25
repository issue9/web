// SPDX-License-Identifier: MIT

// +build linux darwin

//go:generate go build -o=./testdata/plugin_1.so -buildmode=plugin ./testdata/plugin1/plugin.go
//go:generate go build -o=./testdata/plugin_2.so -buildmode=plugin ./testdata/plugin2/plugin.go
//go:generate go build -o=./testdata/plugin3.so -buildmode=plugin ./testdata/plugin3/plugin.go

// Package plugintest 作为插件的功能测试包
//
// NOTE: 该测试如果直接写在功能所在的包，目前版本会报错。
//
// NOTE:其中 plugin3 生成的文件格式与其它不同，用于测试动态加载测试。
package plugintest
