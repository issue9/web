// SPDX-License-Identifier: MIT

package xml

import "github.com/issue9/web/server"

var (
	_ server.MarshalFunc   = Marshal
	_ server.UnmarshalFunc = Unmarshal
)
