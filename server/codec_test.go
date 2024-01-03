// SPDX-License-Identifier: MIT

package server

import (
	"compress/lzw"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/compressor"
	"github.com/issue9/web/mimetype/json"
)

func TestBuildCodec(t *testing.T) {
	a := assert.New(t, false)

	c, err := buildCodec(nil, nil)
	a.NotError(err).NotNil(c)

	c, err = buildCodec(APIMimetypes(), DefaultCompressions())
	a.NotError(err).NotNil(c)

	c, err = buildCodec([]*Mimetype{
		{Name: "application", Marshal: nil, Unmarshal: nil},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
	}, BestCompressionCompressions())
	a.Equal(err.Field, "Mimetypes[1].Name").Nil(c)

	c, err = buildCodec([]*Mimetype{
		{Name: "", Marshal: nil, Unmarshal: nil},
	}, BestSpeedCompressions())
	a.Equal(err.Field, "Mimetypes[0].Name").Nil(c)

	c, err = buildCodec([]*Mimetype{
		{Name: "text", Marshal: nil, Unmarshal: nil},
	}, BestSpeedCompressions())
	a.Equal(err.Field, "Mimetypes[0].Marshal").Nil(c)

	c, err = buildCodec([]*Mimetype{
		{Name: "text", Marshal: json.Marshal, Unmarshal: nil},
	}, BestSpeedCompressions())
	a.Equal(err.Field, "Mimetypes[0].Unmarshal").Nil(c)

	c, err = buildCodec(XMLMimetypes(), []*Compression{
		{Compressor: compressor.NewLZW(lzw.LSB, 8)},
		{Compressor: nil},
	})
	a.Equal(err.Field, "Compressions[1].Compressor").Nil(c)
}
