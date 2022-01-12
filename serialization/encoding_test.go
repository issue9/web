// SPDX-License-Identifier: MIT

package serialization

import (
	"bytes"
	"compress/gzip"
	"log"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestNewEncodings(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		NewEncodings(nil)
	}, "参数 errlog 不能为空")

	a.PanicString(func() {
		NewEncodings(log.Default(), "*")
	}, "无效的值 *")

	e := NewEncodings(log.Default())
	a.NotNil(e)
	a.True(e.allowAny).
		Empty(e.ignoreTypes).
		Empty(e.ignoreTypePrefix)

	e = NewEncodings(log.Default(), "text*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text"})

	e = NewEncodings(log.Default(), "text*", "text/*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text", "text/"})
}

func TestEncodings_Add(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(log.Default())
	a.NotNil(e)

	e.Add(map[string]EncodingWriter{
		"gzip": gzip.NewWriter(&bytes.Buffer{}),
		"br":   gzip.NewWriter(&bytes.Buffer{}),
	})
	a.Equal(2, len(e.algorithms))

	// 重复添加
	a.PanicString(func() {
		e.Add(map[string]EncodingWriter{
			"gzip": gzip.NewWriter(&bytes.Buffer{}),
		})
	}, "存在相同名称的函数")

	a.PanicString(func() {
		e.Add(map[string]EncodingWriter{
			"gzip": nil,
		})
	}, "参数 w 不能为空")

	a.PanicString(func() {
		e.Add(map[string]EncodingWriter{
			"*": gzip.NewWriter(&bytes.Buffer{}),
		})
	}, "name 值不能为 identity 和 *")

	a.PanicString(func() {
		e.Add(map[string]EncodingWriter{
			"identity": gzip.NewWriter(&bytes.Buffer{}),
		})
	}, "name 值不能为 identity 和 *")
}

func TestEncodings_Search(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(log.Default(), "text*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text"})
	e.Add(map[string]EncodingWriter{
		"gzip": gzip.NewWriter(&bytes.Buffer{}),
		"br":   gzip.NewWriter(&bytes.Buffer{}),
	})

	name, w, notAccept := e.Search("application/json", "gzip;q=0.9,br")
	a.False(notAccept).NotNil(w).Equal(name, "br")

	name, w, notAccept = e.Search("application/json", "gzip,br")
	a.False(notAccept).NotNil(w).Equal(name, "gzip")

	name, w, notAccept = e.Search("text/plian", "gzip,br")
	a.False(notAccept).Nil(w).Empty(name)

	// *

	name, w, notAccept = e.Search("application/xml", "*;q=0")
	a.True(notAccept).Nil(w).Empty(name)

	name, w, notAccept = e.Search("application/xml", "*,br")
	a.False(notAccept).NotNil(w).Equal(name, "gzip")

	name, w, notAccept = e.Search("application/xml", "*,gzip")
	a.False(notAccept).NotNil(w).Equal(name, "br")

	name, w, notAccept = e.Search("application/xml", "*,gzip,br")
	a.False(notAccept).Nil(w).Empty(name)

	// identity

	name, w, notAccept = e.Search("application/xml", "identity,gzip,br")
	a.False(notAccept).NotNil(w).Equal(name, "gzip")

	// 正常匹配
	name, w, notAccept = e.Search("application/xml", "identity;q=0,gzip,br")
	a.False(notAccept).NotNil(w).Equal(name, "gzip")

	// 没有可匹配，选取第一个
	name, w, notAccept = e.Search("application/xml", "identity;q=0,abc,def")
	a.False(notAccept).Nil(w).Empty(name)
}
