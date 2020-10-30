// SPDX-License-Identifier: MIT

package validation

// Ruler 验证规则需要实现的接口
type Ruler interface {
	Validate(v interface{}) string
}

// RuleFunc 验证函数的签名
type RuleFunc func(v interface{}) string

// Validate 实现 Ruler.Validate
func (f RuleFunc) Validate(v interface{}) string {
	return f(v)
}
