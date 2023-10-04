// SPDX-License-Identifier: MIT

package json

import "github.com/issue9/web"

var (
	_ web.BuildMarshalFunc = BuildMarshal
	_ web.UnmarshalFunc    = Unmarshal
)
