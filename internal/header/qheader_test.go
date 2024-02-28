// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package header

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestParse(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() {
		ParseQHeader(",a1", "not-allow")
	})

	items := ParseQHeader("a0,a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.8", "*/*")
	a.Length(items, 7)
	// 确定排序是否正常
	a.Equal(items[0].Q, float32(1.0))
	a.Equal(items[6].Q, float32(.5))

	items = ParseQHeader(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.0", "*/*")
	a.Length(items, 6)
	a.Equal(items[0].Q, float32(1.0))

	// xx/* 的权限低于相同 Q 值的其它权限
	items = ParseQHeader("x/*;q=0.1,b/*;q=0.1,a/*;q=0.1,t/*;q=0.1,text/plain;q=0.1", "*/*")
	a.Length(items, 5)
	a.Equal(items[0].Value, "text/plain").Equal(items[0].Q, float32(0.1))
	a.Equal(items[1].Value, "x/*").Equal(items[1].Q, float32(0.1))
	a.Equal(items[2].Value, "b/*").Equal(items[2].Q, float32(0.1))
	a.Equal(items[3].Value, "a/*").Equal(items[3].Q, float32(0.1))
	a.Equal(items[4].Value, "t/*").Equal(items[4].Q, float32(0.1))

	// xx/* 的权限低于相同 Q 值的其它权限
	items = ParseQHeader("text/*;q=0.1,xx/*;q=0.1,text/html;q=0.1", "*/*")
	a.Length(items, 3)
	a.Equal(items[0].Value, "text/html").Equal(items[0].Q, float32(0.1))
	a.Equal(items[1].Value, "text/*").Equal(items[1].Q, float32(0.1))

	// */* 的权限最底
	items = ParseQHeader("text/html;q=0.1,text/*;q=0.1,xx/*;q=0.1,*/*;q=0.1", "*/*")
	a.Length(items, 4)
	a.Equal(items[0].Value, "text/html").Equal(items[0].Q, float32(0.1))
	a.Equal(items[1].Value, "text/*").Equal(items[1].Q, float32(0.1))

	items = ParseQHeader("utf-8;q=x.9,gbk;q=0.8", "*/*")
	a.Length(items, 2)
}

func TestSortItems(t *testing.T) {
	a := assert.New(t, false)

	as := []*Item{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "*/*")

	as = []*Item{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
	}
	sortItems(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "b/*")
	a.Equal(as[2].Value, "*/*")

	as = []*Item{
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

	as = []*Item{
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
	as = []*Item{
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
	as = []*Item{
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

	// 相同 Q 值，Err 不同
	as = []*Item{
		{Value: "zh-cn", Q: 0.7, Err: errors.New("zh-cn")},
		{Value: "zh-tw", Q: 0.8},
		{Value: "*", Q: 0.7},
		{Value: "en", Q: 0.7, Err: errors.New("en")},
		{Value: "en-us", Q: 0},
	}
	sortItems(as, "*")
	a.Equal(as[0].Value, "zh-tw")
	a.Equal(as[1].Value, "*")
	a.Equal(as[2].Value, "en-us") // Q==0，在 Err!=nil 之前
	a.Equal(as[3].Value, "zh-cn")
	a.Equal(as[4].Value, "en")
}
