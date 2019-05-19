// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package versioninfo

const versiongo = `// 由 web release 生成

// Version 版本号
const Version = "%s"

// buildDate 编译日期，可以由编译器指定
var buildDate string

var fullVersion = Version

func init()  {
	if buildDate !=""{
		fullVersion += "+" + buildDate
	}
}

// FullVersion 完整的版本号，可能包括了编译日期。
func FullVersion() string {
	return fullVersion
}
`
