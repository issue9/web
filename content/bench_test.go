// SPDX-License-Identifier: MIT

package content

import (
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		a.True(len(buildContentType(DefaultMimetype, DefaultCharset)) > 0)
	}
}
