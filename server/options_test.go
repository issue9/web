// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/web"
	"golang.org/x/text/language"
)

func TestSanitizeOptions(t *testing.T) {
	a := assert.New(t, false)

	o, err := sanitizeOptions(nil, typeHTTP)
	a.NotError(err).NotNil(o).
		Equal(o.Location, time.Local).
		NotNil(o.Logs).
		NotNil(o.IDGenerator).
		Equal(o.RequestIDKey, header.XRequestID).
		NotNil(o.Negotiation).
		NotZero(len(o.Plugins))

	o, err = sanitizeOptions(&Options{
		Negotiation: web.NewNegotiation(language.MustParse("zh-Hant"), nil),
	}, typeHTTP)
	a.NotError(err).NotNil(o).Equal(o.Negotiation.Language(), language.MustParse("zh-Hant"))
}
