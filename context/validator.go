// SPDX-License-Identifier: MIT

package context

// Validator 验证器
type Validator interface {
	Validate(*Context) map[string][]string
}

// Validate 验证对象的数据
func (ctx *Context) Validate(v interface{}) map[string][]string {
	if vv, ok := v.(Validator); ok {
		return vv.Validate(ctx)
	}
	return nil
}
