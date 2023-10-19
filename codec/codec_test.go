// SPDX-License-Identifier: MIT

package codec

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/config"

	"github.com/issue9/web"
)

var (
	_ web.Codec = &codec{}

	_ config.Sanitizer = &Mimetype{}
	_ config.Sanitizer = &Compression{}
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	c := New(nil, nil)
	a.NotNil(c)

	c = New([]*Mimetype{
		{Name: "application/json", MarshalBuilder: func(*web.Context) web.MarshalFunc { return nil }},
	}, nil)

	a.PanicString(func() {
		c = New([]*Mimetype{
			{Name: "application/json", MarshalBuilder: func(*web.Context) web.MarshalFunc { return nil }},
			{Name: "application/json", MarshalBuilder: func(*web.Context) web.MarshalFunc { return nil }},
		}, nil)
	}, "已经存在同名 application/json 的编码方法")
}

func TestMimetype_SanitizeConfig(t *testing.T) {
	a := assert.New(t, false)

	m := &Mimetype{}
	err := m.SanitizeConfig()
	a.Error(err).Equal(err.Field, "Name")

	m = &Mimetype{Name: "test"}
	err = m.SanitizeConfig()
	a.NotError(err).Equal(m.Problem, m.Name)

	m = &Mimetype{Name: "test", Problem: "p"}
	err = m.SanitizeConfig()
	a.NotError(err).
		Equal(m.Problem, "p").
		Equal(m.Name, "test")
}

func TestCompression_SanitizeConfig(t *testing.T) {
	a := assert.New(t, false)

	c := &Compression{}
	err := c.SanitizeConfig()
	a.Error(err).Equal(err.Field, "Name")

	c = &Compression{Name: "test"}
	err = c.SanitizeConfig()
	a.NotError(err).Equal(c.Types, []string{"*"})

	c = &Compression{Name: "test", Types: []string{"text"}}
	err = c.SanitizeConfig()
	a.NotError(err).Equal(c.Types, []string{"text"})
}
