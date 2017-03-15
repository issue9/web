// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"net/http"

	"github.com/issue9/web/config"
)

var defaultEnvelopeConf = &config.Envelope{
	State:  config.EnvelopeStateDisable,
	Key:    "",
	Status: http.StatusOK,
}
