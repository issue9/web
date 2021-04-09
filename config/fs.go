// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"io/fs"

	"gopkg.in/yaml.v2"
)

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
