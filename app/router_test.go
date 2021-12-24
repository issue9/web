// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestRouter_sanitize(t *testing.T) {
	a := assert.New(t, false)

	r := &router{}
	a.NotError(r.sanitize()).Empty(r.options)

	r = &router{CORS: &cors{}}
	a.NotError(r.sanitize()).Equal(1, len(r.options))

	r = &router{CORS: &cors{Origins: []string{"*"}, AllowCredentials: true}}
	err := r.sanitize()
	a.NotNil(err).Equal(err.Field, "cors.allowCredentials")

	r = &router{CORS: &cors{MaxAge: -2}}
	err = r.sanitize()
	a.NotNil(err).Equal(err.Field, "cors.maxAge")
}
