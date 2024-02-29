// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package comptime 提供了编译期相关的处理方式
//
// go 以 //go:build 的方式区别编译内容，目前支持以下标签：
//   - development 表示开发环境，[Mode] 会被赋为 [Development]；
//   - 其它情况下，则 [Mode] 的值永远是 [Production]；
package comptime

import "github.com/issue9/web"

const (
	Production  int = iota // 运行于生产环境
	Development            // 运行于开发环境
)

// Mode 当前的运行的环境
//
// 这是个编译期的常量，默认情况下始终是 [Production]，
// 只有在编译时指定了 development 标签才会为 [Development]。
const Mode = defaultMode

// Filename 根据当前的环境生成不同的文件名
//
// 按以下规则返回文件名：
//   - [Production] 原样返回；
//   - [Development] 在扩展名前加上 _development，比如 file.yaml => file_development.yaml；
//
// 一般像根据环境加载不同的配置文件之类的功能可以使用此方法。
// 比如 [server/app.CLI.ConfigFilename] 可以使用此文件指定相同的文件名。
func Filename(f string) string { return filename(f) }

// DebugRouter 在 [Development] 环境下为 r 提供一组测试用的 API
//
// path 测试路径；
// id 在取地址参数出错时的 problem id；
func DebugRouter(r *web.Router, path, id string) { debugRouter(r, path, id) }
