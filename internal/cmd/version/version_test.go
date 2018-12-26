// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package version

import (
	"bytes"
	"testing"

	"github.com/issue9/assert"
)

const tags = `14ec481c2d0715306019707b791bb664320e8a7e	refs/tags/v0.16.0
269a60e58b4d69d914e6ce3e701390c830747150	refs/tags/v0.16.1
a07f91201239035ebf85a6423016a6b736b0d037	refs/tags/v0.16.2
974f266a6ae1ff6fa50cee3d124448e3f874704d	refs/tags/v0.16.3
41ee1d50a4545f8a973a65e77056f8d0b6ee6335	refs/tags/v0.16.4
656a3f126dff09c38682aa493ffdbee687600591	refs/tags/v0.17.0
7bae76c94deaeb1fcb4b6f15a69da38edefed803	refs/tags/v0.18.0
53fcf5b89cd51c913e25770fd6cc6e9f53ed0eb7	refs/tags/v0.19.0
3a54a6f661922326859a34ee92ea263bb9b5798c	refs/tags/v0.20.0`

func TestGetMaxVersion(t *testing.T) {
	a := assert.New(t)
	buf := bytes.NewBufferString(tags)

	ver, err := getMaxVersion(buf)
	a.NotError(err).
		Equal(ver, "0.20.0")
}
