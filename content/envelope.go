// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import "net/http"

// Envelope 是否启用的状态
const (
	EnvelopeStateEnable  = iota // 根据客户端决定是否开始
	EnvelopeStateDisable        // 不能使用 envelope
	EnvelopeStateMust           // 只能是 envelope
)

type envelope struct {
	XMLName  struct{}          `json:"-" xml:"xml"`
	Status   int               `json:"status" xml:"status"`
	Headers  map[string]string `json:"headers,omitempty" xml:"headers"`
	Response interface{}       `json:"response,omitempty" xml:"response"`
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
