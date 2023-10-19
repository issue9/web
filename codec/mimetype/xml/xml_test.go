// SPDX-License-Identifier: MIT

package xml

import "github.com/issue9/web"

var (
	_ web.BuildMarshalFunc = BuildMarshal
	_ web.UnmarshalFunc    = Unmarshal
)
