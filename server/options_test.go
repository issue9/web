// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert/v4"
)

func TestSanitizeOptions(t *testing.T) {
	a := assert.New(t, false)

	o, err := sanitizeOptions(nil, typeHTTP)
	a.NotError(err).NotNil(o).
		Equal(o.Location, time.Local).
		NotNil(o.logs).
		NotNil(o.IDGenerator).
		NotNil(o.config).
		Equal(o.Config.Dir, DefaultConfigDir).
		Equal(o.RequestIDKey, RequestIDKey)
}
