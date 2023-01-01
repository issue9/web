// SPDX-License-Identifier: MIT

// Package testdata 测试数据
package testdata

type Object struct {
	Name string `json:"name"`
	Age  int
}

var (
	ObjectInst = &Object{
		Name: "中文",
		Age:  456,
	}

	// {"name":"中文","Age":456}
	ObjectGBKBytes = []byte{'{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 214, 208, 206, 196, '"', ',', '"', 'A', 'g', 'e', '"', ':', '4', '5', '6', '}'}
)

const ObjectJSONString = `{"name":"中文","Age":456}`
