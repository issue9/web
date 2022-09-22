// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/server/servertest"
)

func TestConfigOf_sanitizeEncodings(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{Encodings: []*encodingConfig{
		{Type: "text/*", IDs: []string{"br-default", "compress-msb-8"}},
		{Type: "application/*", IDs: []string{"gzip-default", "compress-msb-8"}},
	}}
	a.NotError(conf.sanitizeEncodings())

	conf = &configOf[empty]{Encodings: []*encodingConfig{
		{Type: "text/*", IDs: []string{"br-default", "compress-msb-8", "not-exists-id"}},
		{Type: "application/*", IDs: []string{"gzip-default", "compress-msb-8"}},
	}}
	err := conf.sanitizeEncodings()
	a.Error(err).Equal(err.Field, "ids")
}

func TestConfigOf_buildEncodings(t *testing.T) {
	a := assert.New(t, false)
	conf := &configOf[empty]{Encodings: []*encodingConfig{
		{Type: "text/*", IDs: []string{"br-default", "compress-msb-8"}},
		{Type: "application/*", IDs: []string{"gzip-default", "compress-msb-8"}},
	}}
	a.NotError(conf.sanitizeEncodings())

	e := servertest.NewServer(a, nil)
	conf.buildEncodings(e.Encodings())
}
