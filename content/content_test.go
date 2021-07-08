// SPDX-License-Identifier: MIT

package content

import (
	"testing"

	"github.com/issue9/assert"
)

func TestBuildContentType(t *testing.T) {
	a := assert.New(t)

	a.Equal("application/xml; charset=utf16", buildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+DefaultCharset, buildContentType("application/xml", ""))
	a.Equal(DefaultMimetype+"; charset="+DefaultCharset, buildContentType("", ""))
	a.Equal("application/xml; charset="+DefaultCharset, buildContentType("application/xml", ""))
}
