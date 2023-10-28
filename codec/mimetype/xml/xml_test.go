// SPDX-License-Identifier: MIT

package xml

import "github.com/issue9/web"

var (
	_ web.MarshalFunc   = Marshal
	_ web.UnmarshalFunc = Unmarshal
)
