// SPDX-License-Identifier: MIT

package problems

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := New("")
	a.NotNil(ps)
	l := len(ps.problems)

	a.False(ps.exists("40010")).
		False(ps.exists("40011"))
	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.True(ps.exists("40010")).
		True(ps.exists("40011")).
		Equal(l+2, len(ps.problems))

	a.PanicString(func() {
		ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	}, "存在相同值的 id 参数")

	a.PanicString(func() {
		ps.Add("40012", 99, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	}, "status 必须是一个有效的状态码")

	a.PanicString(func() {
		ps.Add("40012", 400, nil, nil)
	}, "title 不能为空")
}

func TestProblems_Problem(t *testing.T) {
	a := assert.New(t, false)

	ps := New("")
	a.NotNil(ps)
	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	p := ps.Problem("40010")
	a.Equal(p.Type, "40010").
		Equal(p.id, "40010")

	ps = New("https://example.com/qa#")
	a.NotNil(ps)
	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	p = ps.Problem("40011")
	a.Equal(p.Type, "https://example.com/qa#40011").
		Equal(p.id, "40011")

	ps = New(AboutBlank)
	a.NotNil(ps)
	ps.Add("40012", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	p = ps.Problem("40012")
	a.Equal(p.Type, AboutBlank).
		Equal(p.id, "40012")

	a.PanicString(func() {
		ps.Problem("not-exists")
	}, "未找到有关 not-exists 的定义")
}
