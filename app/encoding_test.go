// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v4"
)

func TestEncodingConfig_build(t *testing.T) {
	a := assert.New(t, false)
	l := logs.New(logs.NewNopWriter())

	var conf *encodingsConfig = nil
	e, err := conf.build(l.ERROR())
	a.NotNil(e).NotError(err)

	conf = &encodingsConfig{}
	e, err = conf.build(l.ERROR())
	a.NotNil(e).NotError(err)
	w, notAccept := e.Search("application/json", "*")
	a.False(notAccept).Nil(w)

	conf = &encodingsConfig{Encodings: []*encodingConfig{{Name: "br", Encoding: "brotli"}}}
	e, err = conf.build(l.ERROR())
	a.NotNil(e).NotError(err)
	w, notAccept = e.Search("application/json", "*")
	a.False(notAccept).NotNil(w)

	conf = &encodingsConfig{Encodings: []*encodingConfig{{Name: "br", Encoding: "br"}}}
	e, err = conf.build(l.ERROR())
	a.Nil(e).Error(err).
		Equal(err.Field, "encodings[br]")
}
