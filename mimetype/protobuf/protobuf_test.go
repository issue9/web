// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package protobuf

import "github.com/issue9/web/mimetype"

var (
	_ mimetype.MarshalFunc   = Marshal
	_ mimetype.UnmarshalFunc = Unmarshal
)
