// SPDX-License-Identifier: MIT

// Package web 一个微型的 RESTful API 框架
package web

import (
	"io/fs"

	"github.com/issue9/logs/v2"
	"github.com/issue9/logs/v2/config"
	"golang.org/x/text/message/catalog"

	cfg "github.com/issue9/web/config"
	"github.com/issue9/web/internal/version"
	"github.com/issue9/web/result"
)

const (
	logsConfigFilename = "logs.xml"
	webconfigFilename  = "web.yaml"
)

// Version 当前框架的版本
func Version() string {
	return version.Version
}

// LoadServer 从配置文件加载并实例化 Server 对象
func LoadServer(name, version string, f fs.FS, c catalog.Catalog, build result.BuildFunc) (*Server, error) {
	conf := &config.Config{}
	if _, err := cfg.Load(logsConfigFilename, conf, cfg.LoadXML(f)); err != nil {
		return nil, err
	}

	l := logs.New()
	if err := l.Init(conf); err != nil {
		return nil, err
	}

	webconfig := &cfg.Webconfig{}
	if _, err := cfg.Load(webconfigFilename, webconfig, cfg.LoadYAML(f)); err != nil {
		return nil, err
	}

	return webconfig.NewServer(name, version, l, c, build)
}
