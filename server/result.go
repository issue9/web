// SPDX-License-Identifier: MIT

package server

import (
	"net/url"
	"strings"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/validation"
	"golang.org/x/text/message"
)

const defaultResultPoolFieldMaxSize = 20

var defaultResultPool = &sync.Pool{New: func() any { return &defaultResult{} }}

type (
	// ResultFields 表示字段的错误信息列表
	//
	// 原始类型为 map[string][]string，键名为字段名，键值为错误信息列表。
	ResultFields = validation.Messages

	// BuildResultFunc 用于生成 Result 接口对象的函数
	//
	// 用户可以通过 BuildResultFunc 返回自定义的 Result 对象，
	// 在 Result 中用户可以自定义其展示方式，可参考默认的实现 DefaultResultBuilder
	BuildResultFunc func(status int, code, message string) Result

	// Result 向客户端输出错误代码时的对象需要实现的接口
	Result interface {
		Responser

		// Add 添加详细的错误信息
		//
		// 一般用于添加具体的字段错误信息，key 表示字段名，val 表示该字段下的具体错误信息。
		Add(key string, val ...string)

		// Set 设置详细的错误信息
		//
		// 与 Add 的区别在于 Set 是覆盖之前的，而 Add 是追加新的。
		Set(key string, val ...string)

		// HasFields 是否存在详细的错误信息
		//
		// 如果有通过 Add 或是 Set 添加内容，那么应该返回 true。
		HasFields() bool
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
// 支持以下格式的返回信息：
//
// JSON:
//
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
//
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
//
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
// protobuf:
//
//  message Result {
//      string message = 1;
//      string code = 2;
//      repeated Field fields = 3;
//  }
//
//  message Field {
//      string name = 1;
//      repeated string message = 2;
//  }
//
// FormData:
//
//  message=errormessage&code=4000001&fields.username=名称过短&fields.username=不能包含特殊符号&fields.password=不能为空
func DefaultResultBuilder(status int, code, message string) Result {
	rslt := defaultResultPool.Get().(*defaultResult)
	rslt.status = status
	rslt.Code = code
	rslt.Message = message
	if len(rslt.Fields) > 0 {
		rslt.Fields = rslt.Fields[:0]
	}
	return rslt
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

func (rslt *defaultResult) Apply(ctx *Context) {
	if err := ctx.Marshal(rslt.status, rslt); err != nil {
		ctx.Logs().ERROR().Error(err)
	}
	if len(rslt.Fields) < defaultResultPoolFieldMaxSize {
		defaultResultPool.Put(rslt)
	}
}

func (rslt *defaultResult) HasFields() bool { return len(rslt.Fields) > 0 }

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

// AddResult 注册错误代码
//
// 此功能与 Server.AddResult 的唯一区别是，code 参数会加上 Module.ID() 作为其前缀。
func (m *Module) AddResult(status int, code string, phrase localeutil.LocaleStringer) {
	m.Server().AddResult(status, m.BuildID(code), phrase)
}

// AddResults 添加多条错误信息
//
// 此功能与 Server.AddResult 的唯一区别是，code 参数会加上 Module.ID() 作为其前缀。
func (m *Module) AddResults(status int, messages map[string]localeutil.LocaleStringer) {
	for k, v := range messages {
		m.AddResult(status, k, v)
	}
}

// NewValidation 声明验证器
//
// separator 用于指定字段名称上下级元素名称之间的连接符。比如在返回 xml 元素时，
// 可能会采用 root/element 的格式表示上下级，此时 separator 应设置为 /。
// 而在 json 中，可能会被转换成 root.element 的格式。
//
// 可以配置 CTXSanitizer 接口使用：
//  type User struct {
//      Name string
//      Age int
//  }
//
//  func(o *User) CTXSanitize(ctx* web.Context) web.ResultFields {
//      v := ctx.NewValidation(".")
//      return v.NewField(o.Name, "name", validator.Required().Message("不能为空")).
//          NewField(o.Age, "age", validator.Min(18).Message("不能小于 18 岁")).
//          Messages()
//  }
//
// 如果需要更加精细的控制，可以调用 github.com/issue9/validation.New 进行声明。
func (ctx *Context) NewValidation(separator string) *validation.Validation {
	return validation.New(validation.ContinueAtError, ctx.LocalePrinter(), separator)
}
