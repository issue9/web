// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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

	a.True(tags[0] > tags[1])
}
