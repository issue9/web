// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

// Envelope 的状态
const (
	EnvelopeStateEnable  = "enable"  // 根据客户端决定是否开始
	EnvelopeStateDisable = "disable" // 不能使用 envelope
	EnvelopeStateMust    = "must"    // 只能是 envelope
)

// Envelope 配置
type Envelope struct {
	State  string `json:"state"`
	Key    string `json:"key"`
	Status int    `json:"status"`
}
