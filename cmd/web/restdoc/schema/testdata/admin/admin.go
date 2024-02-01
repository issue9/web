// SPDX-License-Identifier: MIT

// Package admin 测试用例
package admin

import (
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/schema/testdata"
)

// User testdata.User
type User testdata.User

type Alias = testdata.User

type State = web.State

type Sex = testdata.Sex

type Admin struct {
	XMLName struct{} `xml:"admin"`

	testdata.User                  // User
	U1            []*testdata.User // u1
	U2            testdata.User    `json:"u2,omitempty"` // u2
	u3            Alias            // 不可导出
	U4            User
}

type IntUserGenerics = testdata.Generics[int, User]

type Generics[T1 any, T2 any] struct {
	G1 testdata.Generics[T1, T2]
}

type Int64UserGenerics = Generics[int64, User]
