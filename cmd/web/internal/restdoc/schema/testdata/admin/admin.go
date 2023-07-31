// SPDX-License-Identifier: MIT

// Package admin 测试用例
package admin

import "github.com/issue9/web/cmd/web/internal/restdoc/schema/testdata"

type User testdata.User

type Admin struct {
	testdata.User                  // User
	U1            []*testdata.User // u1
	U2            testdata.User    `json:"u2,omitempty"` // u2
	u3            testdata.User
	U4            User
}

type IntStringGenerics = testdata.Generics[int, Admin]
