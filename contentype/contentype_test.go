// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package contentype

import (
	"github.com/issue9/web/contentype/internal/json"
	"github.com/issue9/web/contentype/internal/xml"
)

var _ ContentTyper = json.New(nil)

var _ ContentTyper = xml.New(nil)
