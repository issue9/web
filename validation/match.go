// SPDX-License-Identifier: MIT

package validation

import "regexp"

// Match 定义正则匹配的验证规则
func Match(msg string, exp *regexp.Regexp) Ruler {
	return RuleFunc(func(v interface{}) (ret string) {
		var ok bool
		switch vv := v.(type) {
		case string:
			ok = exp.MatchString(vv)
		case []byte:
			ok = exp.Match(vv)
		case []rune:
			ok = exp.MatchString(string(vv))
		default:
			return msg
		}

		if !ok {
			ret = msg
		}
		return
	})
}
