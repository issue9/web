// SPDX-License-Identifier: MIT

package content

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/issue9/validation"
)

type (
	// Fields 表示字段的错误信息列表
	//
	// 原始类型为 map[string][]string
	Fields = validation.Messages

	// BuildResultFunc 用于生成 Result 接口对象的函数
	BuildResultFunc func(status, code int, message string) Result

	// Result 自定义错误代码的实现接口
	//
	// 一般是对客户端提交数据 400 的具体反馈信息。
	// 用户可以根据自己的需求，展示自定义的错误码以及相关的错误信息格式。
	// 该对象最终也是调用 MarshalFunc 进行解码输出。
	// 只要该对象同时实现了 Result 接口即可。
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

	defaultResult struct {
		XMLName struct{} `json:"-" xml:"result" yaml:"-"`

		status int // 当前的信息所对应的 HTTP 状态码

		Message string         `json:"message" xml:"message" yaml:"message"`
		Code    int            `json:"code" xml:"code,attr" yaml:"code"`
		Fields  []*fieldDetail `json:"fields,omitempty" xml:"field,omitempty" yaml:"fields,omitempty"`
	}

	fieldDetail struct {
		Name    string   `json:"name" xml:"name,attr" yaml:"name"`
		Message []string `json:"message" xml:"message" yaml:"message"`
	}
)

// DefaultBuilder 默认的 BuildResultFunc 实现
//
// 定义了以下格式的返回信息：
//
// JSON:
//  {
//      'message': 'error message',
//      'code': 4000001,
//      'fields':[
//          {'name': 'username': 'message': ['名称过短', '不能包含特殊符号']},
//          {'name': 'password': 'message': ['不能为空']},
//      ]
//  }
//
// XML:
//  <result code="400">
//      <message>error message</message>
//      <field name="username">
//          <message>名称过短</message>
//          <message>不能包含特殊符号</message>
//      </field>
//      <field name="password"><message>不能为空</message></field>
//  </result>
//
// YAML:
//  message: 'error message'
//  code: 40000001
//  fields:
//    - name: username
//      message:
//        - 名称过短
//        - 不能包含特殊符号
//    - name: password
//      message:
//        - 不能为空
//
// FormData:
//  message=errormessage&code=4000001&fields.username=名称过短&fields.username=不能包含特殊符号&fields.password=不能为空
func DefaultBuilder(status, code int, message string) Result {
	return &defaultResult{
		status:  status,
		Code:    code,
		Message: message,
	}
}

func (rslt *defaultResult) Add(field string, message ...string) {
	for _, d := range rslt.Fields {
		if d.Name == field {
			d.Message = append(d.Message, message...)
			return
		}
	}
	rslt.Fields = append(rslt.Fields, &fieldDetail{Name: field, Message: message})
}

func (rslt *defaultResult) Set(field string, message ...string) {
	for _, d := range rslt.Fields {
		if d.Name == field {
			d.Message = message
			return
		}
	}
	rslt.Fields = append(rslt.Fields, &fieldDetail{Name: field, Message: message})
}

func (rslt *defaultResult) Status() int {
	return rslt.status
}

func (rslt *defaultResult) HasFields() bool {
	return len(rslt.Fields) > 0
}

func (rslt *defaultResult) MarshalForm() ([]byte, error) {
	vals := url.Values{}

	vals.Add("code", strconv.Itoa(rslt.Code))
	vals.Add("message", rslt.Message)

	for _, field := range rslt.Fields {
		k := "fields." + field.Name
		for _, msg := range field.Message {
			vals.Add(k, msg)
		}
	}

	return []byte(vals.Encode()), nil
}

func (rslt *defaultResult) UnmarshalForm(b []byte) error {
	vals, err := url.ParseQuery(string(b))
	if err != nil {
		return err
	}

	for key, vals := range vals {
		switch key {
		case "code":
			if rslt.Code, err = strconv.Atoi(vals[0]); err != nil {
				return err
			}
		case "message":
			rslt.Message = vals[0]
		default:
			name := strings.TrimPrefix(key, "fields.")
			rslt.Fields = append(rslt.Fields, &fieldDetail{
				Name:    name,
				Message: vals,
			})
		}
	}

	return nil
}
