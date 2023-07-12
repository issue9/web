// SPDX-License-Identifier: MIT

// Package testdata 测试 schema 的生成
package testdata

type (
	String string

	// Sex 表示性别
	// @enum female male unknown
	Sex int8

	// 用户信息 doc
	User struct { // 用户信息 comment
		// 姓名
		Name String

		// 年龄
		Age int `xml:"age,attr" json:"age"`
		Sex Sex `json:"sex" xml:"sex,attr"` // 性别
	}
)
