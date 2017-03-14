// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
)

// Envelope 状态下返回的状态值
const envelopeStatus = http.StatusOK

type envelope struct {
	Status   int               `json:"status"`
	Headers  map[string]string `json:"headers"`
	Response interface{}       `json:"response"`
}

func newEnvelope(code int, headers http.Header, resp interface{}) *envelope {
	hs := make(map[string]string, len(headers))
	for key := range headers {
		hs[key] = headers.Get(key)
	}

	return &envelope{
		Status:   code,
		Headers:  hs,
		Response: resp,
	}
}
