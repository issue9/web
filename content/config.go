// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"errors"
	"net/http"

	"github.com/issue9/utils"
)

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

// DefaultConfig 获取一个默认的 Config 实例。
func DefaultConfig() *Config {
	return &Config{
		ContentType:    "json",
		EnvelopeState:  EnvelopeStateDisable,
		EnvelopeKey:    "envelope",
		EnvelopeStatus: http.StatusOK,
	}
}

// Init 初始化配置内容，给空值赋予默认址。
func (conf *Config) Init() error {
	return utils.Merge(true, conf, DefaultConfig())
}

// Check 检测各个项的各法性
func (conf *Config) Check() error {
	if conf.ContentType != "json" && conf.ContentType != "xml" {
		return errors.New("contentType 无效的值")
	}

	if conf.EnvelopeState != EnvelopeStateMust &&
		conf.EnvelopeState != EnvelopeStateEnable &&
		conf.EnvelopeState != EnvelopeStateDisable {
		return errors.New("envelopeState 无效的值")
	}

	if conf.EnvelopeState != EnvelopeStateDisable {
		switch {
		case len(conf.EnvelopeKey) == 0:
			return errors.New("envelopeKey 不能为空")
		case conf.EnvelopeStatus < 100 || conf.EnvelopeStatus > 999:
			return errors.New("envelopeStatus 值无效")
		}
	}
	return nil
}
