// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v4"

	"github.com/issue9/web/serialization"
)

func TestEncodingConfig_build(t *testing.T) {
	a := assert.New(t, false)
	l := logs.New(logs.NewNopWriter())

	var conf *encodingsConfig = nil
	e := conf.build(l.ERROR())
	a.NotNil(e)

	conf = &encodingsConfig{}
	e = conf.build(l.ERROR())
	a.NotNil(e)
	w, notAccept := e.Search("application/json", "*/*")
	a.False(notAccept).Nil(w)

	conf = &encodingsConfig{Encodings: map[string]serialization.EncodingWriterFunc{"br": serialization.BrotliWriter}}
	e = conf.build(l.ERROR())
	a.NotNil(e)
	w, notAccept = e.Search("application/json", "*/*")
	a.False(notAccept).NotNil(w)
}
