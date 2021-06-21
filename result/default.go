// SPDX-License-Identifier: MIT

package result

type defaultResult struct {
	XMLName struct{} `json:"-" xml:"result" yaml:"-" form:"-"`

	status int // 当前的信息所对应的 HTTP 状态码

	Message string         `json:"message" xml:"message" yaml:"message" form:"message"`
	Code    int            `json:"code" xml:"code,attr" yaml:"code" form:"code"`
	Fields  []*fieldDetail `json:"fields,omitempty" xml:"field,omitempty" yaml:"fields,omitempty" form:"fields"`
}

type fieldDetail struct {
	Name    string   `json:"name" xml:"name,attr" yaml:"name" form:"name"`
	Message []string `json:"message" xml:"message" yaml:"message" form:"message"`
}

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
