// SPDX-License-Identifier: MIT

package context

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/text/message"
)

// BuildResultFunc 用于生成 Result 接口对象的函数
type BuildResultFunc func(status, code int, message string) Result

type resultMessage struct {
	message string
	status  int // 对应的 HTTP 状态码
}

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
//
// 可以在 default.go 查看 Result 的实现方式。
type Result interface {
	// 添加详细的内容
	Add(key, val string)

	// 设置详细的内容
	Set(key, val string)

	// 是否存在详细的错误信息
	//
	// 如果有通过 Add 添加内容，那么应该返回 true
	HasFields() bool

	// HTTP 状态码
	//
	// 最终会经此值作为 HTTP 状态会返回给用户
	Status() int
}

// CTXResult Result 与 Context 相结合的实现
type CTXResult struct {
	rslt Result
	ctx  *Context
}

// 这是对 Result 的默认实现
type defaultResult struct {
	XMLName struct{} `json:"-" xml:"result" yaml:"-"`

	status int // 当前的信息所对应的 HTTP 状态码

	Message string         `json:"message" xml:"message" yaml:"message"`
	Code    int            `json:"code" xml:"code,attr" yaml:"code"`
	Fields  []*fieldDetail `json:"fields,omitempty" xml:"field,omitempty" yaml:"fields,omitempty"`
}

type fieldDetail struct {
	Name    string   `json:"name" xml:"name,attr" yaml:"name"`
	Message []string `json:"message" xml:"message" yaml:"message"`
}

// Messages 错误信息列表
//
// p 用于返回特定语言的内容。如果为空，则表示返回原始值。
func (b *Builder) Messages(p *message.Printer) map[int]string {
	msgs := make(map[int]string, len(b.messages))

	if p == nil {
		for code, msg := range b.messages {
			msgs[code] = msg.message
		}
	} else {
		for code, msg := range b.messages {
			msgs[code] = p.Sprintf(msg.message)
		}
	}

	return msgs
}

// AddMessages 添加一组错误信息。
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
// msgs 中，键名表示的是该错误的错误代码；
// 键值表示具体的错误描述内容。
func (b *Builder) AddMessages(status int, msgs map[int]string) {
	for code, msg := range msgs {
		if msg == "" {
			panic("参数 msg 不能为空值")
		}

		if _, found := b.messages[code]; found {
			panic(fmt.Sprintf("重复的消息 ID: %d", code))
		}

		b.messages[code] = &resultMessage{message: msg, status: status}
	}
}

// NewResult 查找指定代码的错误信息
func (b *Builder) NewResult(code int) Result {
	msg, found := b.messages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	return b.build(msg.status, code, msg.message)
}

// NewResult 返回 CTXResult 实例
func (ctx *Context) NewResult(code int) *CTXResult {
	return &CTXResult{
		rslt: ctx.builder.NewResult(code),
		ctx:  ctx,
	}
}

// NewResultWithFields 返回 CTXResult 实例
func (ctx *Context) NewResultWithFields(code int, detail map[string]string) *CTXResult {
	rslt := ctx.NewResult(code)

	for k, v := range detail {
		rslt.Add(k, v)
	}

	return rslt
}

// Add 添加详细的内容
func (rslt *CTXResult) Add(key, val string) *CTXResult {
	rslt.rslt.Add(key, val)
	return rslt
}

// HasFields 是否存在详细的错误信息
//
// 如果有通过 Add 添加内容，那么应该返回 true
func (rslt *CTXResult) HasFields() bool {
	return rslt.rslt.HasFields()
}

// Render 渲染内容
func (rslt *CTXResult) Render() {
	rslt.ctx.Render(rslt.rslt.Status(), rslt.rslt, nil)
}

// DefaultResultBuilder 默认的 BuildResultFunc 实现
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
func DefaultResultBuilder(status, code int, message string) Result {
	return &defaultResult{
		status:  status,
		Code:    code,
		Message: message,
	}
}

func (rslt *defaultResult) Add(field, message string) {
	for _, d := range rslt.Fields {
		if d.Name == field {
			d.Message = append(d.Message, message)
			return
		}
	}
	rslt.Fields = append(rslt.Fields, &fieldDetail{Name: field, Message: []string{message}})
}

func (rslt *defaultResult) Set(field, message string) {
	for _, d := range rslt.Fields {
		if d.Name == field {
			d.Message = d.Message[:1]
			d.Message[0] = message
			return
		}
	}
	rslt.Fields = append(rslt.Fields, &fieldDetail{Name: field, Message: []string{message}})
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
