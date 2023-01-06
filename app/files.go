// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"
	"io/fs"
	"path/filepath"
	"strconv"

	"github.com/issue9/localeutil"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/errs"
	"github.com/issue9/web/internal/files"
)

var filesFactory = map[string]files.FileSerializer{}

type MarshalFileFunc = files.MarshalFunc

type UnmarshalFileFunc = files.UnmarshalFunc

func (conf *configOf[T]) sanitizeFiles() *errs.ConfigError {
	conf.files = make(map[string]files.FileSerializer, len(conf.Files))
	for i, name := range conf.Files {
		s, found := filesFactory[name]
		if !found {
			return errs.NewConfigError("["+strconv.Itoa(i)+"]", localeutil.Phrase("not found serialization function for %s", name))
		}
		conf.files[name] = s // conf.Files 可以保证 conf.files 唯一性
	}
	return nil
}

// RegisterFileSerializer 注册用于文件序列化的方法
//
// ext 为文件的扩展名，如果存在同名，则会覆盖。
func RegisterFileSerializer(m MarshalFileFunc, u UnmarshalFileFunc, ext ...string) {
	for _, e := range ext {
		filesFactory[e] = files.FileSerializer{Marshal: m, Unmarshal: u}
	}
}

func loadConfigOf[T any](fsys fs.FS, path string) (*configOf[T], error) {
	ext := filepath.Ext(path)
	var item files.FileSerializer
	for name, ss := range filesFactory {
		if name == ext {
			item = ss
			break
		}
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	conf := &configOf[T]{}
	if err := item.Unmarshal(data, conf); err != nil {
		return nil, err
	}

	if err := conf.sanitize(); err != nil {
		err.Path = path
		return nil, err
	}

	return conf, nil
}

func init() {
	RegisterFileSerializer(json.Marshal, json.Unmarshal, ".json")
	RegisterFileSerializer(xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
}
