// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type textObject struct {
	Name string
	Age  int
}

func (o *textObject) MarshalText() ([]byte, error) {
	return []byte(o.Name + "," + strconv.Itoa(o.Age)), nil
}

func (o *textObject) UnmarshalText(data []byte) error {
	text := strings.Split(string(data), ",")
	if len(text) != 2 {
		return errors.New("无法转换")
	}

	age, err := strconv.Atoi(text[1])
	if err != nil {
		return err
	}
	o.Age = age
	o.Name = text[0]
	return nil
}

func TestAddMarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(v interface{}) ([]byte, error) { return nil, nil }

	a.Equal(1, len(marshals))

	a.NotError(AddMarshal("n1", f1))
	a.NotError(AddMarshal("n2", f1))
	a.Equal(AddMarshal("n2", f1), ErrExists)
	a.Equal(3, len(marshals))
}

func TestAddUnmarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(data []byte, v interface{}) error { return nil }

	a.Equal(1, len(unmarshals))

	a.NotError(AddUnmarshal("n1", f1))
	a.NotError(AddUnmarshal("n2", f1))
	a.Equal(AddUnmarshal("n2", f1), ErrExists)
	a.Equal(3, len(unmarshals))
}

func TestAddCharset(t *testing.T) {
	a := assert.New(t)
	e := simplifiedchinese.GBK

	a.Equal(1, len(charset))
	a.Nil(charset[DefaultCharset])

	a.NotError(AddCharset("n1", e))
	a.NotError(AddCharset("n2", e))
	a.Equal(AddCharset("n2", e), ErrExists)
	a.Equal(3, len(charset))
}

func TestTextMarshal(t *testing.T) {
	a := assert.New(t)

	v := "123"
	data, err := textMarshal(v)
	a.NotError(err).Equal(string(data), v)

	data, err = textMarshal([]rune(v))
	a.NotError(err).Equal(string(data), v)

	// 实现 TextMarshaler 的对象
	data, err = textMarshal(&textObject{Name: "test", Age: 5})
	a.NotError(err).NotNil(data).Equal(string(data), "test,5")

	// 未实现 TextMarshaler 接口的对象
	data, err = textMarshal(&struct{}{})
	a.Equal(err, errUnsupportedMarshal).Nil(data)
}

func TestTextUnmarshal(t *testing.T) {
	a := assert.New(t)

	v1 := &textObject{}
	a.NotError(textUnmarshal([]byte("test,5"), v1))
	a.Equal(v1.Name, "test").Equal(v1.Age, 5)

	v2 := &struct{}{}
	a.Equal(textUnmarshal(nil, v2), errUnsupportedMarshal)
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t)

	a.Equal("application/xml; charset=utf16", buildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+DefaultCharset, buildContentType("application/xml", ""))
	a.Equal(DefaultEncoding+"; charset="+DefaultCharset, buildContentType("", ""))
	a.Equal("application/xml; charset="+DefaultCharset, buildContentType("application/xml", ""))
}

func TestParseContentType(t *testing.T) {
	a := assert.New(t)

	e, c := parseContentType("")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = parseContentType(" ")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = parseContentType(" ;;;")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = parseContentType("application/XML")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType("application/XML;")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType(" application/xml ;;")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType("  ;charset=utf16")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")

	e, c = parseContentType("  ;charset=utf16;")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")

	e, c = parseContentType("application/xml; charset=utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = parseContentType("application/xml;utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = parseContentType("application/xml;=utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = parseContentType("application/xml;utf16=")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType("application/xml;=utf16=")
	a.Equal(e, "application/xml").Equal(c, "utf16=")

	e, c = parseContentType(";charset=utf16")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")
}
