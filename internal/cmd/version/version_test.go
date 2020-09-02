// SPDX-License-Identifier: MIT

package version

import (
	"testing"

	"github.com/issue9/assert"
)

func init() {
	repoURL = "./"
}

func TestGetRemoteTags(t *testing.T) {
	a := assert.New(t)

	tags, err := getRemoteTags()
	a.NotError(err).NotNil(tags)

	if len(tags) > 2 {
		a.True(tags[0] > tags[1])
	}
}
