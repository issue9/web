// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestNewEncodings(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		NewEncodings(nil, "*")
	}, "无效的值 *")

	e := NewEncodings(nil)
	a.NotNil(e)
	a.True(e.allowAny).
		Empty(e.ignoreTypes).
		Empty(e.ignoreTypePrefix)

	e = NewEncodings(nil, "text*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text"})

	e = NewEncodings(nil, "text*", "text/*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text", "text/"})
}

func TestEncodings_Add(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add(map[string]NewEncodingFunc{
		"gzip": GZipWriter,
		"br":   BrotliWriter,
	})
	a.Equal(2, len(e.pools))

	// 重复添加
	a.PanicString(func() {
		e.Add(map[string]NewEncodingFunc{
			"gzip": GZipWriter,
		})
	}, "存在相同名称的函数")

	a.PanicString(func() {
		e.Add(map[string]NewEncodingFunc{
			"gzip": nil,
		})
	}, "参数 w 不能为空")

	a.PanicString(func() {
		e.Add(map[string]NewEncodingFunc{
			"*": GZipWriter,
		})
	}, "name 值不能为 identity 和 *")

	a.PanicString(func() {
		e.Add(map[string]NewEncodingFunc{
			"identity": GZipWriter,
		})
	}, "name 值不能为 identity 和 *")
}

func TestEncodings_Search(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil, "text*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text"})
	e.Add(map[string]NewEncodingFunc{
		"gzip": GZipWriter,
		"br":   GZipWriter,
	})

	b, notAccept := e.Search("application/json", "gzip;q=0.9,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "br")

	b, notAccept = e.Search("application/json", "gzip,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

	b, notAccept = e.Search("text/plian", "gzip,br")
	a.False(notAccept).Nil(b)

	b, notAccept = e.Search("text/plian", "")
	a.False(notAccept).Nil(b)

	// *

	b, notAccept = e.Search("application/xml", "*;q=0")
	a.True(notAccept).Nil(b)

	b, notAccept = e.Search("application/xml", "*,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

	b, notAccept = e.Search("application/xml", "*,gzip")
	a.False(notAccept).NotNil(b).Equal(b.name, "br")

	b, notAccept = e.Search("application/xml", "*,gzip,br")
	a.False(notAccept).Nil(b)

	// identity

	b, notAccept = e.Search("application/xml", "identity,gzip,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

	// 正常匹配
	b, notAccept = e.Search("application/xml", "identity;q=0,gzip,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

	// 没有可匹配，选取第一个
	b, notAccept = e.Search("application/xml", "identity;q=0,abc,def")
	a.False(notAccept).Nil(b)
}

func TestEncodings_Compress(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil, "text*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text"})
	e.Add(map[string]NewEncodingFunc{
		"gzip": GZipWriter,
		"br":   GZipWriter,
	})

	b, notAccept := e.Search("application/json", "gzip;q=0.9,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "br")

	w := &bytes.Buffer{}
	wc := b.Get(w)
	_, err := wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	r, err := gzip.NewReader(w)
	a.NotError(err).NotNil(r)
	data, err := io.ReadAll(r)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")
}
