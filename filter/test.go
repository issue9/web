// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package filter

import "github.com/issue9/localeutil"

// Test 测试过滤器并返回错误信息
//
// exitAtError 出现错误时是否直接跳过其它执行其它过滤器；
// p 用于转换错误信息的本地化信息；
func Test(exitAtError bool, p *localeutil.Printer, f ...Filter) map[string]string {
	if len(f) == 0 {
		return nil
	}

	errs := make(map[string]string, len(f))
	for _, ff := range f {
		if name, msg := ff(); msg != nil {
			errs[name] = msg.LocaleString(p)

			if exitAtError {
				break
			}
		}
	}

	return errs
}
