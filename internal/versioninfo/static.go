// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package versioninfo

const versiongo = `// 由工具自动生成，不能修改！

package version

// Version 版本号
//
// 版本号规则遵循 https://semver.org/lang/zh-CN/
const Version = "%s"

// 编译日期，可以由编译器指定
var %s string

// 最后一次提交的 hash 值
var %s string

var fullVersion = Version

func init()  {
	if %s !=""{
		fullVersion = Version + "+" + %s
	}
}

// FullVersion 完整的版本号，可能包括了编译日期。
func FullVersion() string {
	return fullVersion
}

// CommitHash 最后一次提示我的 hash 值
func CommitHash() string {
	return %s
}
`
