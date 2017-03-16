// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

// Envelope 的状态
const (
	EnvelopeStateEnable  = "enable"  // 根据客户端决定是否开始
	EnvelopeStateDisable = "disable" // 不能使用 envelope
	EnvelopeStateMust    = "must"    // 只能是 envelope
)

// Config 初始化 content 包的配置
type Config struct {
	ContentType    string `json:"contentType"` // 默认的编码类型
	EnvelopeState  string `json:"envelopeState"`
	EnvelopeKey    string `json:"envelopeKey"`
	EnvelopeStatus int    `json:"envelopeStatus"`
}
