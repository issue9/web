// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"errors"
	"net/http"
)

// Envelope 的状态
//
// Envelope 为用户端在无法获取报头等内容的情况下，
// 服务端将这些内容转换成一定格式输出到响应内容中的一种方式。
const (
	EnvelopeStateEnable  = "enable"  // 根据客户端决定是否开始
	EnvelopeStateDisable = "disable" // 不能使用 envelope
	EnvelopeStateMust    = "must"    // 强制使用 envelope
)

// Config New 的参数，同时也是 internal/config 中配置文件中的部分内容。
type Config struct {
	ContentType    string `json:"contentType"`    // 编码类型
	EnvelopeState  string `json:"envelopeState"`  // envelope 的启用状态
	EnvelopeKey    string `json:"envelopeKey"`    // 若 EnvelopeState == enable，则此值表示触发的查询参数关键字。
	EnvelopeStatus int    `json:"envelopeStatus"` // 当 Envelope 处理开启状态时，返回的状态码。
}

// DefaultConfig 获取一个默认的 Config 实例。
func DefaultConfig() *Config {
	return &Config{
		ContentType:    "json",
		EnvelopeState:  EnvelopeStateDisable,
		EnvelopeKey:    "envelope",
		EnvelopeStatus: http.StatusOK,
	}
}

// Sanitize 检测各项的各法性
func (conf *Config) Sanitize() error {
	if conf.ContentType != "json" && conf.ContentType != "xml" {
		return errors.New("contentType 无效的值")
	}

	if conf.EnvelopeState != EnvelopeStateMust &&
		conf.EnvelopeState != EnvelopeStateEnable &&
		conf.EnvelopeState != EnvelopeStateDisable {
		return errors.New("envelopeState 无效的值")
	}

	if conf.EnvelopeState == EnvelopeStateEnable {
		switch {
		case len(conf.EnvelopeKey) == 0:
			return errors.New("envelopeKey 不能为空")
		case conf.EnvelopeStatus < 100 || conf.EnvelopeStatus > 599:
			return errors.New("envelopeStatus 值无效")
		}
	}

	if conf.EnvelopeState == EnvelopeStateMust &&
		(conf.EnvelopeStatus < 100 || conf.EnvelopeStatus > 599) {
		return errors.New("envelopeStatus 值无效")
	}
	return nil
}
