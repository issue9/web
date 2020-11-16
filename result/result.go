// SPDX-License-Identifier: MIT

// Package result 提供对自定义错误代码的支持
package result

import "github.com/issue9/validation"

type (
	// Fields 表示字段的错误信息列表
	//
	// 类型为 map[string][]string
	Fields = validation.Messages

	// BuildFunc 用于生成 Result 接口对象的函数
	BuildFunc func(status, code int, message string) Result

	// Result 自定义错误代码的实现接口
	//
	// 用户可以根据自己的需求，在出错时，展示自定义的错误码以及相关的错误信息格式。
	// 只要该对象实现了 Result 接口即可。
	//
	// 比如类似以下的错误内容：
	//  {
	//      'message': 'error message',
	//      'code': 4000001,
	//      'detail':[
	//          {'field': 'username': 'message': '已经存在相同用户名'},
	//          {'field': 'username': 'message': '已经存在相同用户名'},
	//      ]
	//  }
	Result interface {
		// 添加详细的错误信息
		//
		// 相同的 key 应该能关联多个 val 值。
		Add(key string, val ...string)

		// 设置详细的错误信息
		//
		// 如果已经相同的 key，会被覆盖。
		Set(key string, val ...string)

		// 是否存在详细的错误信息
		//
		// 如果有通过 Add 添加内容，那么应该返回 true
		HasFields() bool

		// HTTP 状态码
		//
		// 最终会经此值作为 HTTP 状态会返回给用户
		Status() int
	}
)
