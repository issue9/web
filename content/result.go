// SPDX-License-Identifier: MIT

package content

import (
	"fmt"
	"net/url"
	"strconv"
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

	resultMessage struct {
		status int
		localeutil.LocaleStringer
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

func (rslt *defaultResult) Status() int { return rslt.status }

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

// Results 错误信息列表
//
// p 用于返回特定语言的内容。
func (c *Content) Results(p *message.Printer) map[int]string {
	msgs := make(map[int]string, len(c.resultMessages))
	for code, msg := range c.resultMessages {
		msgs[code] = msg.LocaleString(p)
	}
	return msgs
}

// AddResults 添加多条错误信息
//
// 键名为错误信息的数字代码，键值是具体的错误信息描述。
// 同时键名向下取整，直到三位长度的整数作为其返回给客户端的状态码。
func (c *Content) AddResults(messages map[int]localeutil.LocaleStringer) {
	for code, phrase := range messages {
		status := calcStatus(code)
		c.AddResult(status, code, phrase)
	}
}

// AddResult 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
func (c *Content) AddResult(status, code int, phrase localeutil.LocaleStringer) {
	if _, found := c.resultMessages[code]; found {
		panic(fmt.Sprintf("重复的消息 ID: %d", code))
	}
	c.resultMessages[code] = &resultMessage{status: status, LocaleStringer: phrase}
}

// Result 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
// fields 表示明细字段，可以为空，之后通过 Result.Add 添加。
func (c *Content) Result(p *message.Printer, code int, fields ResultFields) Result {
	msg, found := c.resultMessages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	rslt := c.resultBuilder(msg.status, code, msg.LocaleString(p))
	for k, vals := range fields {
		rslt.Add(k, vals...)
	}

	return rslt
}

func calcStatus(code int) int {
	if code < 1000 {
		panic("无效的 code")
	}

	status := code / 10
	for status > 999 {
		status = status / 10
	}
	return status
}
