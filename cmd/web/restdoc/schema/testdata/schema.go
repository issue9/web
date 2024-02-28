// SPDX-FileCopyrightText: 2018-2024 caixw
//
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
//
// @enum female male unknown
// @type string
type Sex int8

type Sexes []Sex

// 测试自动获取枚举类型
// @type string
// @enums
type Enum int

const (
	EnumV1 Enum = iota + 1
	EnumV2

	EnumV3
)

// @enum
type NotBasicTypeEnum struct{}

const (
	E1 NotBasicTypeEnum = struct{}
)

// @type number
// @enum 1 2
type Number int

const (
	V1 Number = iota
	V2
)
