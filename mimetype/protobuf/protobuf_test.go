// SPDX-License-Identifier: MIT

package protobuf

import "github.com/issue9/web/mimetype"

var (
	_ mimetype.MarshalFunc   = Marshal
	_ mimetype.UnmarshalFunc = Unmarshal
)
