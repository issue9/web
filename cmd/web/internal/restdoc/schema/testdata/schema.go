// SPDX-License-Identifier: MIT

// Package testdata 测试 schema 的生成
package testdata

import "time"

type (
	String string

	// 用户信息 doc
	User struct { // 用户信息 comment
		// 姓名
		Name String

		// 年龄
		Age int `xml:"age,attr" json:"age"`
		Sex Sex `json:"sex" xml:"sex,attr"` // 性别

		// struct doc
		Struct struct {
			// x doc
			X int // x comment
		} `json:"struct"` // struct comment

		Birthday time.Time
	}

	Generic[T any] struct {
		Type T
	}

	IntGeneric = Generic[int]

	Generics[T1 any, T2 any] struct {
		F1 Generic[T1]
		F2 *T2
		P  int
	}
)

// Sex 表示性别
// @enum female male unknown
// @type string
type Sex int8

type Sexes []Sex
