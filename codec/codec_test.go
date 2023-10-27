// SPDX-License-Identifier: MIT

package codec

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
	"github.com/issue9/web/codec/compressor"
	"github.com/issue9/web/locales"
)

var _ web.Codec = &codec{}

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	c, err := New("", "", nil, nil)
	a.NotError(err).NotNil(c)

	c, err = New("ms", "cs", []*Mimetype{
		{Name: "application/json", MarshalBuilder: func(*web.Context) web.MarshalFunc { return nil }},
	}, BestSpeedCompressions())
	a.NotError(err).NotNil(c)

	c, err = New("ms", "cs", []*Mimetype{
		{Name: "application/json", MarshalBuilder: func(*web.Context) web.MarshalFunc { return nil }},
		{Name: "application/json", MarshalBuilder: func(*web.Context) web.MarshalFunc { return nil }},
	}, nil)
	a.Equal(err.Message, locales.DuplicateValue).Nil(c).
		Equal(err.Field, "ms[0].Name")

	c, err = New("ms", "cs", []*Mimetype{
		{Name: "", MarshalBuilder: func(*web.Context) web.MarshalFunc { return nil }},
	}, nil)
	a.Equal(err.Field, "ms[0].Name").Nil(c)

	c, err = New("ms", "cs", nil, []*Compression{{}})
	a.Equal(err.Field, "cs[0].Compressor").Nil(c)
}

func TestMimetype_sanitize(t *testing.T) {
	a := assert.New(t, false)

	m := &Mimetype{}
	err := m.sanitize()
	a.Error(err).Equal(err.Field, "Name")

	m = &Mimetype{Name: "test"}
	err = m.sanitize()
	a.NotError(err).Equal(m.Problem, m.Name)

	m = &Mimetype{Name: "test", Problem: "p"}
	err = m.sanitize()
	a.NotError(err).
		Equal(m.Problem, "p").
		Equal(m.Name, "test")
}

func TestCompression_sanitize(t *testing.T) {
	a := assert.New(t, false)

	c := &Compression{}
	err := c.sanitize()
	a.Error(err).Equal(err.Field, "Compressor")

	c = &Compression{Compressor: compressor.NewZstdCompressor()}
	err = c.sanitize()
	a.NotError(err).
		True(c.wildcard).
		Length(c.Types, 0).
		Length(c.wildcardSuffix, 0)

	c = &Compression{Compressor: compressor.NewZstdCompressor(), Types: []string{"text"}}
	err = c.sanitize()
	a.NotError(err).Equal(c.Types, []string{"text"})
}
