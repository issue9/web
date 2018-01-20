// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding"
)

// 返回一个默认的 Config
func defaultConfig() *config {
	return &config{
		Debug:          true,
		OutputCharset:  encoding.DefaultCharset,
		OutputEncoding: encoding.DefaultEncoding,
		Strict:         true,

		HTTPS:          false,
		Domain:         "example.com",
		HTTPState:      httpStateDisabled,
		CertFile:       "",
		KeyFile:        "",
		Port:           8082,
		Headers:        nil,
		Static:         nil,
		DisableOptions: false,
		AllowedDomains: []string{},

		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,

		Marshals: map[string]encoding.Marshal{
			encoding.DefaultEncoding: encoding.TextMarshal,
		},
		Unmarshals: map[string]encoding.Unmarshal{
			encoding.DefaultEncoding: encoding.TextUnmarshal,
		},
		Charset: map[string]encoding.Charset{
			encoding.DefaultCharset: nil,
		},
	}
}

func TestConfig_sanitize(t *testing.T) {
	a := assert.New(t)

	// 不存在 allowedDomains，不会将 Domain 加入其中
	conf := defaultConfig()
	conf.Domain = "example.com"
	conf.AllowedDomains = nil
	a.NotError(conf.sanitize())
	a.Nil(conf.AllowedDomains)

	// 存在 allowedDomains，将 Domain 加入其中
	conf = defaultConfig()
	conf.Domain = "example.com"
	conf.AllowedDomains = []string{"caixw.io"}
	a.NotError(conf.sanitize())
	a.Equal(2, len(conf.AllowedDomains))

	// 存在 allowedDomains 且有与 Domain 相同的项，不会再次将 Domain 加入其中
	conf = defaultConfig()
	conf.Domain = "example.com"
	conf.AllowedDomains = []string{"caixw.io", "example.com"}
	a.NotError(conf.sanitize())
	a.Equal(2, len(conf.AllowedDomains))
}
