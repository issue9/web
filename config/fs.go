// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"io/fs"

	"github.com/issue9/logs/v2"
	"github.com/issue9/logs/v2/config"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/result"
	"github.com/issue9/web/server"
)

const (
	logsConfigFilename = "logs.xml"
	webconfigFilename  = "web.yaml"
)

// NewServer 从配置文件初始化 Server 实例
func NewServer(name, version string, f fs.FS, c catalog.Catalog, b result.BuildFunc) (*server.Server, error) {
	conf := &config.Config{}
	if _, err := Load(logsConfigFilename, conf, LoadXML(f)); err != nil {
		return nil, err
	}

	l := logs.New()
	if err := l.Init(conf); err != nil {
		return nil, err
	}

	webconfig := &Webconfig{}
	if _, err := Load(webconfigFilename, webconfig, LoadYAML(f)); err != nil {
		return nil, err
	}

	return webconfig.buildServer(name, version, f, l, c, b)
}

// LoadYAML 加载 YAML 的配置文件并转换成 v 对象的内容
func LoadYAML(f fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		return yaml.Unmarshal(data, v)
	}
}

// LoadJSON 加载 JSON 的配置文件并转换成 v 对象的内容
func LoadJSON(f fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, v)
	}
}

// LoadXML 加载 XML 的配置文件并转换成 v 对象的内容
func LoadXML(f fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		return xml.Unmarshal(data, v)
	}
}
