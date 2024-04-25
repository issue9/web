// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v8/header"

	"github.com/issue9/web/mimetype/html"
)

var _ html.Marshaler = &RenderResponse{}

func TestSanitizeOptions(t *testing.T) {
	a := assert.New(t, false)

	o, err := sanitizeOptions(nil, typeHTTP)
	a.NotError(err).NotNil(o).
		Equal(o.Location, time.Local).
		NotNil(o.Logs).
		NotNil(o.IDGenerator).
		NotNil(o.Config).
		Equal(o.RequestIDKey, header.XRequestID).
		NotNil(o.locale).
		NotNil(o.Codec).
		NotZero(len(o.Plugins))
}
