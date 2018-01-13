// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

// 返回一个默认的 Config
func defaultConfig() *config {
	return &config{
		Debug:          true,
		OutputCharset:  "utf-8",
		OutputEncoding: "application/json",
		Strict:         true,

		HTTPS:          false,
		HTTPState:      httpStateDisabled,
		CertFile:       "",
		KeyFile:        "",
		Port:           httpPort,
		Headers:        nil,
		Static:         nil,
		Options:        true,
		Version:        "",
		AllowedDomains: []string{},

		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

func TestLoad(t *testing.T) {
	a := assert.New(t)

	conf, err := loadConfig("./testdata/web.yaml")
	a.NotError(err).NotNil(conf)
	a.Equal(conf.Root, "https://caixw.io")
	a.Equal(conf.KeyFile, "keyFile")
	a.Equal(conf.ReadTimeout, 3*time.Second)
}
