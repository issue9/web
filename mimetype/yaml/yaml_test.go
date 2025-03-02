// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package yaml

import "github.com/issue9/web"

var (
	_ web.MarshalFunc   = Marshal
	_ web.UnmarshalFunc = Unmarshal
)
