// SPDX-License-Identifier: MIT

package server

import (
	"net/url"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/validation"
	"golang.org/x/text/message"
)

type (
	// ResultFields 表示字段的错误信息列表
	//
	// 原始类型为 map[string][]string
	ResultFields = validation.Messages

	// BuildResultFunc 用于生成 Result 接口对象的函数
	//
	// 用户可以通过 BuildResultFunc 返回自定义的 Result 对象，
	// 在 Result 中用户可以自定义其展示方式，可参考默认的实现 DefaultResultBuilder
	BuildResultFunc func(status int, code, message string) Result

	// Result 展示错误代码需要实现的接口
	Result interface {
		// Add 添加详细的错误信息
		//
		// 相同的 key 应该能关联多个 val 值。
		Add(key string, val ...string)

		// Set 设置详细的错误信息
		//
		// 如果已经相同的 key，会被覆盖。
		Set(key string, val ...string)

		// HasFields 是否存在详细的错误信息
		//
		// 如果有通过 Add 添加内容，那么应该返回 true
		HasFields() bool

		// Status HTTP 状态码
		//
		// 最终会经此值作为 HTTP 状态会返回给用户
		Status() int
	}

	defaultResult struct {
		XMLName struct{} `json:"-" xml:"result" yaml:"-"`

		status int // 当前的信息所对应的 HTTP 状态码

		Message string         `json:"message" xml:"message" yaml:"message"`
		Code    string         `json:"code" xml:"code,attr" yaml:"code"`
		Fields  []*fieldDetail `json:"fields,omitempty" xml:"field,omitempty" yaml:"fields,omitempty"`
	}

	fieldDetail struct {
		Name    string   `json:"name" xml:"name,attr" yaml:"name"`
		Message []string `json:"message" xml:"message" yaml:"message"`
	}

	resultMessage struct {
		status int
		localeutil.LocaleStringer
	}
)

// DefaultResultBuilder 默认的 BuildResultFunc 实现
//
// 定义了以下格式的返回信息：
//
// JSON:
//  {
//      'message': 'error message',
//      'code': '4000001',
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
//  code: '40000001'
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
func DefaultResultBuilder(status int, code, message string) Result {
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

func (rslt *defaultResult) Status() int { return rslt.status }

func (rslt *defaultResult) HasFields() bool {
	return len(rslt.Fields) > 0
}

func (rslt *defaultResult) MarshalForm() ([]byte, error) {
	vals := url.Values{}

	vals.Add("code", rslt.Code)
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
			rslt.Code = vals[0]
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

// Results 返回错误代码以及对应的说明内容
func (srv *Server) Results(p *message.Printer) map[string]string {
	msgs := make(map[string]string, len(srv.resultMessages))
	for code, msg := range srv.resultMessages {
		msgs[code] = msg.LocaleString(p)
	}
	return msgs
}

// AddResults 添加多条错误信息
func (srv *Server) AddResults(status int, messages map[string]localeutil.LocaleStringer) {
	for code, phrase := range messages {
		srv.AddResult(status, code, phrase)
	}
}

// AddResult 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
func (srv *Server) AddResult(status int, code string, phrase localeutil.LocaleStringer) {
	if _, found := srv.resultMessages[code]; found {
		panic("重复的消息 ID: " + code)
	}
	srv.resultMessages[code] = &resultMessage{status: status, LocaleStringer: phrase}
}

// Result 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
// fields 表示明细字段，可以为空，之后通过 Result.Add 添加。
func (srv *Server) Result(p *message.Printer, code string, fields ResultFields) Result {
	msg, found := srv.resultMessages[code]
	if !found {
		panic("不存在的错误代码: " + code)
	}

	rslt := srv.resultBuilder(msg.status, code, msg.LocaleString(p))
	for k, vals := range fields {
		rslt.Add(k, vals...)
	}

	return rslt
}
