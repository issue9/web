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

const errInfoPoolFieldMaxSize = 20

var errInfoPool = &sync.Pool{New: func() any { return &errInfo{} }}

type (
	// CTXSanitizer 提供对数据的验证和修正
	//
	// 在 Context.Read 和 Queries.Object 中会在解析数据成功之后，调用该接口进行数据验证。
	CTXSanitizer interface {
		// CTXSanitize 验证和修正当前对象的数据
		//
		// 如果验证有误，则需要返回这些错误信息。
		CTXSanitize(*Context) FieldErrs
	}

	// FieldErrs 表示字段的错误信息列表
	//
	// 原始类型为 map[string][]string，键名为字段名，键值为错误信息列表。
	FieldErrs = validation.LocaleMessages

	// BuildErrInfoFunc 返回自定义的错误对象
	//
	// 用户可以自定义其展示方式，可参考默认的实现 DefaultErrInfoBuilder
	BuildErrInfoFunc func(status int, code, message string, fields FieldErrs) Responser

	errInfo struct {
		XMLName struct{} `json:"-" xml:"errors" yaml:"-"`

		status int // 当前的信息所对应的 HTTP 状态码

		Message string      `json:"message" xml:"message" yaml:"message"`
		Code    string      `json:"code" xml:"code,attr" yaml:"code"`
		Fields  []*errField `json:"fields,omitempty" xml:"field,omitempty" yaml:"fields,omitempty"`
	}

	errField struct {
		Name    string   `json:"name" xml:"name,attr" yaml:"name"`
		Message []string `json:"message" xml:"message" yaml:"message"`
	}

	errMessage struct {
		status int
		localeutil.LocaleStringer
	}
)

// DefaultErrInfoBuilder 默认的 BuildErrInfoFunc 实现
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
//  <errors code="400">
//      <message>error message</message>
//      <field name="username">
//          <message>名称过短</message>
//          <message>不能包含特殊符号</message>
//      </field>
//      <field name="password"><message>不能为空</message></field>
//  </errors>
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
//  message Errors {
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
//  message=error message&code=4000001&fields.username=名称过短&fields.username=不能包含特殊符号&fields.password=不能为空
func DefaultErrInfoBuilder(status int, code, message string, fields FieldErrs) Responser {
	rslt := errInfoPool.Get().(*errInfo)
	rslt.status = status
	rslt.Code = code
	rslt.Message = message
	if len(rslt.Fields) > 0 {
		rslt.Fields = rslt.Fields[:0]
	}
	for k, v := range fields {
		rslt.Fields = append(rslt.Fields, &errField{Name: k, Message: v})
	}
	return rslt
}

func (rslt *errInfo) Apply(ctx *Context) {
	if err := ctx.Marshal(rslt.status, rslt); err != nil {
		ctx.Logs().ERROR().Error(err)
	}
	if len(rslt.Fields) < errInfoPoolFieldMaxSize {
		errInfoPool.Put(rslt)
	}
}

func (rslt *errInfo) MarshalForm() ([]byte, error) {
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

func (rslt *errInfo) UnmarshalForm(b []byte) error {
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
			rslt.Fields = append(rslt.Fields, &errField{
				Name:    name,
				Message: vals,
			})
		}
	}

	return nil
}

// ErrInfos 返回错误代码以及对应的说明内容
func (srv *Server) ErrInfos(p *message.Printer) map[string]string {
	msgs := make(map[string]string, len(srv.errInfo))
	for code, msg := range srv.errInfo {
		msgs[code] = msg.LocaleString(p)
	}
	return msgs
}

// AddErrInfos 添加多条错误信息
func (srv *Server) AddErrInfos(status int, messages map[string]localeutil.LocaleStringer) {
	for code, phrase := range messages {
		srv.AddErrInfo(status, code, phrase)
	}
}

// AddErrInfo 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
func (srv *Server) AddErrInfo(status int, code string, phrase localeutil.LocaleStringer) {
	if _, found := srv.errInfo[code]; found {
		panic("重复的消息 ID: " + code)
	}
	srv.errInfo[code] = &errMessage{status: status, LocaleStringer: phrase}
}

// ErrInfo 返回 ErrInfo 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
// fields 表示明细字段，可以为空，之后通过 ErrInfo.Add 添加。
func (srv *Server) ErrInfo(p *message.Printer, code string, fields FieldErrs) Responser {
	if msg, found := srv.errInfo[code]; found {
		return srv.errInfoBuilder(msg.status, code, msg.LocaleString(p), fields)
	}
	panic("不存在的错误代码: " + code)
}

// AddErrInfo 注册错误代码
//
// 此功能与 Server.AddErrInfo 的唯一区别是，code 参数会加上 Module.ID() 作为其前缀。
func (m *Module) AddErrInfo(status int, code string, phrase localeutil.LocaleStringer) {
	m.Server().AddErrInfo(status, m.BuildID(code), phrase)
}

// AddErrInfos 添加多条错误信息
//
// 此功能与 Server.AddErrInfos 的唯一区别是，code 参数会加上 Module.ID() 作为其前缀。
func (m *Module) AddErrInfos(status int, messages map[string]localeutil.LocaleStringer) {
	for k, v := range messages {
		m.AddErrInfo(status, k, v)
	}
}

// NewValidation 声明验证器
//
// 一般配合 CTXSanitizer 接口使用：
//
//  type User struct {
//      Name string
//      Age int
//  }
//
//  func(o *User) CTXSanitize(ctx* web.Context) web.FieldErrs {
//      v := ctx.NewValidation(10)
//      return v.NewField(o.Name, "name", validator.Required().Message("不能为空")).
//          NewField(o.Age, "age", validator.Min(18).Message("不能小于 18 岁")).
//          LocaleMessages(ctx.localePrinter())
//  }
//
// cap 表示为错误信息预分配的大小；
func (ctx *Context) NewValidation(cap int) *validation.Validation {
	return validation.New(validation.ContinueAtError, cap)
}
