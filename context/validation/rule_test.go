// SPDX-License-Identifier: MIT

package validation

var _ Ruler = RuleFunc(func(interface{}) string { return "" })
