// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"github.com/issue9/errwrap"
	"github.com/issue9/localeutil"
)

// Collector 过滤器的错误信息收集函数
type Collector func(name string, reason localeutil.Stringer)

// Map [Collector] 收集的错误信息的存取接口
//
// 该对象可由 [CreateMapCollector] 创建。
type Map map[string][]localeutil.Stringer

func (f Map) collect(name string, reason localeutil.Stringer) {
	if s, found := f[name]; found {
		f[name] = append(s, reason)
	} else {
		f[name] = []localeutil.Stringer{reason}
	}
}

// CreateMapCollector 创建一个以 map 形式保存的 [Collector] 实现
//
// 返回的 Map 用于存取已经收集的错误信息，Collector 则可以传递给 [New] 的 collector 参数。
func CreateMapCollector() (Map, Collector) {
	f := Map{}
	return f, f.collect
}

func (m Map) Error() string { return m.LocaleString(nil) }

func (m Map) LocaleString(p *localeutil.Printer) string {
	if len(m) == 0 {
		return ""
	}

	sb := errwrap.StringBuilder{}

	for k, v := range m {
		sb.WString(k).WByte(':')
		for _, vv := range v {
			sb.WString(vv.LocaleString(p)).WByte(';')
		}
		sb.WByte('\n')
	}

	return sb.String()
}
