// SPDX-License-Identifier: MIT

package result

import (
	"net/url"
	"strconv"
	"strings"
)

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

func (rslt *defaultResult) HasDetail() bool {
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
