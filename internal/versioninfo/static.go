// SPDX-License-Identifier: MIT

package versioninfo

const versionGo = `// 当前文件由 https://github.com/issue9/web 自动生成，请不要手动修改！

// Package version 程序的版本信息
package version

import (
	"strconv"
	"strings"
	"time"

	_ "embed"
)

//go:embed VERSION
var versionString string

var info *Info

// Info 版本的相关信息
type Info struct {
	Raw string // 原始的版本字符串
	Main string // 主版本
	Date time.Time // 编译时间
	Hash string // git 提交的 hash 值
	Commits int // 最的次提交的相对于 tag 的提交数量
}

func init() {
	info =&Info{Raw:versionString}

	if index :=strings.IndexByte(versionString, '+');index > 0 {
		info.Main = versionString[:index]
		versionString = versionString[index+1:]
	}

	
	if index := strings.IndexByte(versionString, '.');index > 0 {
		date,err := time.Parse("20060102",versionString[:index])
		if err != nil {
			panic(err)
		}
		info.Date=date

		versionString = versionString[index+1:]
	}


	if index := strings.IndexByte(versionString, '.');index > 0 {
		num,err :=strconv.Atoi(versionString[:index])
		if err!=nil{
			panic(err)
		}
		info.Commits = num

		info.Hash = versionString[index+1:]
	}
}

//  Version 返回版本号
func Version() *Info{
	return info
}
`
