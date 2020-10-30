// SPDX-License-Identifier: MIT

package validation

// IfExpr 根据 if 条件选择不同的验证规则
type IfExpr struct {
	expr      bool
	ifRules   []Ruler
	elseRules []Ruler
}

// If 返回 IfExpr 表达式
func If(expr bool, rule ...Ruler) *IfExpr {
	return &IfExpr{
		expr:    expr,
		ifRules: rule,
	}
}

// Else 指定条件不成言的验证规则
//
// 调用多次，则以最后一次指定为准，如果最后一次为空，则取消 Else 的内容。
func (expr *IfExpr) Else(rule ...Ruler) *IfExpr {
	expr.elseRules = rule
	return expr
}

// Rules 返回当前表达式最后使用的验证规则
func (expr *IfExpr) Rules() []Ruler {
	if expr.expr {
		return expr.ifRules
	}
	return expr.elseRules
}
