// SPDX-License-Identifier: MIT

package json

import "github.com/issue9/web/server"

var (
	_ server.MarshalFunc   = Marshal
	_ server.UnmarshalFunc = Unmarshal
)
