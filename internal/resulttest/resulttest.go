// SPDX-License-Identifier: MIT

// Package resulttest 提供了 app.Result 接口的默认实现，方便测试用。
package resulttest

// New 返回 Result 对象
func New(status, code int, message string) *Result {
	return &Result{
		status:  status,
		Code:    code,
		Message: message,
	}
}

// Result 定义了出错时，向客户端返回的结构体。支持以下格式：
//
// JSON:
//  {
//      'message': 'error message',
//      'code': 4000001,
//      'detail':[
//          {'field': 'username': 'message': '已经存在相同用户名'},
//          {'field': 'username': 'message': '已经存在相同用户名'},
//      ]
//  }
//
// XML:
//  <result code="400" message="error message">
//      <field name="username">已经存在相同用户名</field>
//      <field name="username">已经存在相同用户名</field>
//  </result>
//
// YAML:
//  message: 'error message'
//  code: 40000001
//  detail:
//    - field: username
//      message: 已经存在相同用户名
//    - field: username
//      message: 已经存在相同用户名
//
// FormData:
//  message=errormessage&code=4000001&detail.username=message&detail.username=message
type Result struct {
	XMLName struct{} `json:"-" xml:"result" yaml:"-"`

	// 当前的信息所对应的 HTTP 状态码
	status int

	Message string    `json:"message" xml:"message,attr" yaml:"message" protobuf:"bytes,2,opt,name=message,proto3"`
	Code    int       `json:"code" xml:"code,attr" yaml:"code" protobuf:"varint,1,opt,name=code,proto3"`
	Detail  []*detail `json:"detail,omitempty" xml:"field,omitempty" yaml:"detail,omitempty" protobuf:"bytes,3,rep,name=detail,proto3"`
}

type detail struct {
	Field   string `json:"field" xml:"name,attr" yaml:"field" protobuf:"bytes,1,opt,name=field,proto3"`
	Message string `json:"message" xml:",chardata" yaml:"message" protobuf:"bytes,2,opt,name=message,proto3"`
}

// Add app.Result.Add
func (rslt *Result) Add(field, message string) {
	rslt.Detail = append(rslt.Detail, &detail{Field: field, Message: message})
}

// Status app.Result.Status
func (rslt *Result) Status() int {
	return rslt.status
}

// HasDetail app.Result.Status
func (rslt *Result) HasDetail() bool {
	return len(rslt.Detail) > 0
}

// Error app.Result.Error
func (rslt *Result) Error() string {
	return rslt.Message
}

// Reset proto.Message.Reset
func (rslt *Result) Reset() {
	*rslt = Result{}
}
