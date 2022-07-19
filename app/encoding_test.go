// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/web/internal/encoding"
)

func TestConfigOf_sanitizeEncodings(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{Encodings: []*encodingConfig{
		{Name: "text/*", IDs: []string{"br-default", "compress-msb-8"}},
		{Name: "application/*", IDs: []string{"gzip-default", "compress-msb-8"}},
	}}
	a.NotError(conf.sanitizeEncodings())

	conf = &configOf[empty]{Encodings: []*encodingConfig{
		{Name: "text/*", IDs: []string{"br-default", "compress-msb-8", "not-exists-id"}},
		{Name: "application/*", IDs: []string{"gzip-default", "compress-msb-8"}},
	}}
	err := conf.sanitizeEncodings()
	a.Error(err).Equal(err.Field, "ids")
}

func TestConfigOf_buildEncodings(t *testing.T) {
	a := assert.New(t, false)
	conf := &configOf[empty]{Encodings: []*encodingConfig{
		{Name: "text/*", IDs: []string{"br-default", "compress-msb-8"}},
		{Name: "application/*", IDs: []string{"gzip-default", "compress-msb-8"}},
	}}
	a.NotError(conf.sanitizeEncodings())

	e := encoding.NewEncodings(nil)
	a.NotError(conf.buildEncodings(e))
}
