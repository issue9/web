// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package form

import "github.com/issue9/web/encoding"

var (
	_ encoding.MarshalFunc   = Marshal
	_ encoding.UnmarshalFunc = Unmarshal
)
