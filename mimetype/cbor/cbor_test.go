// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package cbor

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/mimetypetest"
)

var (
	_ web.MarshalFunc   = Marshal
	_ web.UnmarshalFunc = Unmarshal
)

func TestCBOR(t *testing.T) {
	a := assert.New(t, false)
	mimetypetest.Test(a, Marshal, Unmarshal)
}
