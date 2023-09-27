// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestConfigOf_sanitizeCompresses(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{Encodings: []*compressConfig{
		{Types: []string{"text/*", "application/*"}, ID: "compress-msb-8"},
		{Types: []string{"text/*"}, ID: "br-default"},
		{Types: []string{"application/*"}, ID: "gzip-default"},
	}}
	a.NotError(conf.sanitizeEncodings())

	conf = &configOf[empty]{Encodings: []*compressConfig{
		{Types: []string{"text/*"}, ID: "compress-msb-8"},
		{Types: []string{"text/*"}, ID: "not-exists-id"},
	}}
	err := conf.sanitizeEncodings()
	a.Error(err).Equal(err.Field, "compresses[1].id")
}
