// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert"
)

func TestRouter_sanitize(t *testing.T) {
	a := assert.New(t)

	r := &Router{}
	a.NotError(r.sanitize()).Empty(r.options)

	r = &Router{CORS: &CORS{}}
	a.NotError(r.sanitize()).Equal(1, len(r.options))

	r = &Router{CORS: &CORS{Origins: []string{"*"}, AllowCredentials: true}}
	err := r.sanitize()
	a.NotNil(err).Equal(err.Field, "cors.allowCredentials")

	r = &Router{CORS: &CORS{MaxAge: -2}}
	err = r.sanitize()
	a.NotNil(err).Equal(err.Field, "cors.maxAge")
}
