// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	stdjson "encoding/json"
	"net/http"
)

// Envelope 是否启用的状态
const (
	envelopeStateEnable  = iota // 根据客户端决定是否开始
	envelopeStateDisable        // 不能使用 envelope
	envelopeStateMust           // 只能是 envelope
)

type envelope struct {
	XMLName  struct{}    `json:"-" xml:"xml"`
	Status   int         `json:"status" xml:"status"`
	Headers  []*header   `json:"headers,omitempty" xml:"headers>header,omitempty"`
	Response interface{} `json:"response,omitempty" xml:"response,omitempty"`
}

type header struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

func (h *header) MarshalJSON() ([]byte, error) {
	return stdjson.Marshal(map[string]string{h.Name: h.Value})
}

func newEnvelope(code int, headers http.Header, resp interface{}) *envelope {
	hs := make([]*header, 0, len(headers))
	for key := range headers {
		hs = append(hs, &header{Name: key, Value: headers.Get(key)})
	}

	return &envelope{
		Status:   code,
		Headers:  hs,
		Response: resp,
	}
}
