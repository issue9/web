// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
)

func TestSanitizeOptions(t *testing.T) {
	a := assert.New(t, false)

	o, err := sanitizeOptions(nil, typeHTTP)
	a.NotError(err).NotNil(o)
	a.Equal(o.Location, time.Local).
		NotNil(o.logs).
		NotNil(o.IDGenerator).
		NotNil(o.config).
		Equal(o.Config.Dir, DefaultConfigDir)

	a.Equal(o.RequestIDKey, RequestIDKey)
}
