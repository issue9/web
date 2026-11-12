// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package qheader

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestParse(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() {
		ParseQHeader(",a1", "not-allow")
	})

	items, deny, err := ParseQHeader("a0,a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.8", "*/*")
	a.Empty(err).
		Empty(deny).
		Length(items, 7).
		Equal(items[0].Q, float32(1.0)).
		Equal(items[6].Q, float32(.5))

	items, deny, err = ParseQHeader(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.8", "*/*")
	a.Empty(err).
		Empty(deny).
		Length(items, 6).
		Equal(items[0].Q, float32(1.0))

	// 舍弃 q==0 的
	items, deny, err = ParseQHeader(",a1,a2,a3;q=0,a4,a5;q=0.0,a6", "*/*")
	a.Empty(err).
		Equal(deny, []string{"a3", "a5"}).
		Length(items, 4).
		Equal(items[0].Q, float32(1.0))

	// */*;q=0 的
	items, deny, err = ParseQHeader(",a1,a2,a3;q=0,a4,*/*;q=0.0,a6", "*/*")
	a.Empty(err).
		Equal(deny, []string{"a3", "*/*"}).
		Length(items, 4).
		Equal(items[0].Q, float32(1.0))

	// xx/* 的权限低于相同 Q 值的其它权限
	items, deny, err = ParseQHeader("x/*;q=0.1,b/*;q=0.1,a/*;q=0.1,t/*;q=0.1,text/plain;q=0.1", "*/*")
	a.Empty(err).
		Empty(deny).
		Length(items, 5).
		Equal(items[0].Value, "text/plain").Equal(items[0].Q, float32(0.1)).
		Equal(items[1].Value, "x/*").Equal(items[1].Q, float32(0.1)).
		Equal(items[2].Value, "b/*").Equal(items[2].Q, float32(0.1)).
		Equal(items[3].Value, "a/*").Equal(items[3].Q, float32(0.1)).
		Equal(items[4].Value, "t/*").Equal(items[4].Q, float32(0.1))

	// xx/* 的权限低于相同 Q 值的其它权限
	items, deny, err = ParseQHeader("text/*;q=0.1,xx/*;q=0.1,text/html;q=0.1", "*/*")
	a.Empty(err).
		Empty(deny).
		Length(items, 3).
		Equal(items[0].Value, "text/html").Equal(items[0].Q, float32(0.1)).
		Equal(items[1].Value, "text/*").Equal(items[1].Q, float32(0.1))

	// */* 的权限最底
	items, deny, err = ParseQHeader("text/html;q=0.1,text/*;q=0.1,xx/*;q=0.1,*/*;q=0.1", "*/*")
	a.Empty(err).
		Empty(deny).
		Length(items, 4).
		Equal(items[0].Value, "text/html").Equal(items[0].Q, float32(0.1)).
		Equal(items[1].Value, "text/*").Equal(items[1].Q, float32(0.1))

	items, deny, err = ParseQHeader("utf-8;q=x.9,gbk;q=0.8", "*/*")
	a.Equal(err, []string{"utf-8;q=x.9"}).
		Empty(deny).
		Length(items, 1)

	items, deny, err = ParseQHeader("aaa;q=0.9, bbb;q=0.5", "*/*")
	a.Empty(err).
		Length(items, 2).
		Empty(deny).
		Equal(items[0].Value, "aaa").
		Equal(items[1].Value, "bbb")

	items, deny, err = ParseQHeader("xxx, , yyy", "*/*")
	a.Empty(err).
		Length(items, 2).
		Empty(deny).
		Equal(items[0].Value, "xxx").
		Equal(items[1].Value, "yyy")

	// 空字符串
	items, deny, err = ParseQHeader("", "*/*")
	a.Empty(err).
		Empty(deny).
		Equal(items[0].Value, "*/*")
}

func TestSortItems(t *testing.T) {
	a := assert.New(t, false)

	as := []Item{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "*/*")

	as = []Item{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "b/*")
	a.Equal(as[2].Value, "*/*")

	as = []Item{
		{Value: "*/*", Q: 0.8},
		{Value: "a/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "*/*")
	a.Equal(as[1].Value, "a/*")
	a.Equal(as[2].Value, "b/*")

	as = []Item{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
		{Value: "c/c", Q: 0.7},
		{Value: "b/*", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "c/c")
	a.Equal(as[1].Value, "a/*")
	a.Equal(as[2].Value, "b/*")
	a.Equal(as[3].Value, "*/*")

	as = []Item{
		{Value: "d/d", Q: 0.7},
		{Value: "a/*", Q: 0.7},
		{Value: "*/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
		{Value: "c/c", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "d/d")
	a.Equal(as[1].Value, "c/c")
	a.Equal(as[2].Value, "a/*")
	a.Equal(as[3].Value, "b/*")
	a.Equal(as[4].Value, "*/*")

	// Q 值不一样
	as = []Item{
		{Value: "d/d", Q: 0.7},
		{Value: "a/*", Q: 0.8},
		{Value: "*/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
		{Value: "c/c", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "d/d")
	a.Equal(as[2].Value, "c/c")
	a.Equal(as[3].Value, "b/*")
	a.Equal(as[4].Value, "*/*")

	// 相同 Q 值，保持原样
	as = []Item{
		{Value: "zh-cn", Q: 0.7},
		{Value: "zh-tw", Q: 0.8},
		{Value: "*", Q: 0.7},
		{Value: "en", Q: 0.7},
		{Value: "en-us", Q: 0.7},
	}
	sortItems(as, "*")
	a.Equal(as[0].Value, "zh-tw")
	a.Equal(as[1].Value, "zh-cn")
	a.Equal(as[2].Value, "en")
	a.Equal(as[3].Value, "en-us")
	a.Equal(as[4].Value, "*")
}
