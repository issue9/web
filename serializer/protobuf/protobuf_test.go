// SPDX-License-Identifier: MIT

package protobuf

import "github.com/issue9/web/serializer"

var (
	_ serializer.MarshalFunc   = Marshal
	_ serializer.UnmarshalFunc = Unmarshal
)
