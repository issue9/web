// SPDX-License-Identifier: MIT

package protobuf

import "github.com/issue9/web/serialization"

var (
	_ serialization.MarshalFunc   = Marshal
	_ serialization.UnmarshalFunc = Unmarshal
)
