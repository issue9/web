// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/fs"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

// FS 基于文件系统的配置项管理
type FS struct {
	fs       fs.FS
	selector SelectorFunc
}

// NewFS 声明一个空的 FS 对象
func NewFS(fs fs.FS) *FS {
	return &FS{
		fs: fs,
	}
}

// EncodingSelector 返回常用编码的选择器
//
// 提供了对 json、yaml 和 xml 的支持。
func EncodingSelector(fs *FS) SelectorFunc {
	return func(name string) UnmarshalFunc {
		ext := strings.ToLower(path.Ext(name))
		switch ext {
		case ".json":
			return LoadJSON(fs.fs)
		case ".xml":
			return LoadXML(fs.fs)
		case ".yaml", ".yml":
			return LoadYAML(fs.fs)
		default:
			return nil
		}
	}
}

// Selector 设置选择器
func (f *FS) Selector(selector func(string) UnmarshalFunc) {
	f.selector = selector
}

// Load Loader.Load 接口方法实现
func (f *FS) Load(name string, v interface{}) (Refresher, error) {
	u := f.selector(name)
	if u == nil {
		return nil, errors.New("无法处理的文档类型")
	}
	return Load(name, v, u), nil
}

// LoadYAML 加载 YAML 的配置文件并转换成 v 对象的内容
func LoadYAML(fsys fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		return yaml.Unmarshal(data, v)
	}
}

// LoadJSON 加载 JSON 的配置文件并转换成 v 对象的内容
func LoadJSON(fsys fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, v)
	}
}

// LoadXML 加载 XML 的配置文件并转换成 v 对象的内容
func LoadXML(fsys fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		return xml.Unmarshal(data, v)
	}
}
