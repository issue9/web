// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var errExtNotAllowEmpty = errors.New("扩展名不能为空")

// Sanitizer 如果对象实现了该方法，那么在解析完之后，
// 会调用该接口的函数对数据进行修正和检测。
type Sanitizer interface {
	Sanitize() error
}

// Manager 对 UnmarshalFunc 与扩展名的关联管理。
type Manager struct {
	dir        string
	unmarshals map[string]UnmarshalFunc
}

// NewManager 新的 Manager 实例
func NewManager(dir string) (*Manager, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return &Manager{
		dir:        dir,
		unmarshals: map[string]UnmarshalFunc{},
	}, nil
}

// Exists 是否已经存在该类型的解析函数
//
// ext 表示该类型的文件扩展名，带 . 符号
func (mgr *Manager) Exists(ext string) bool {
	_, found := mgr.unmarshals[ext]
	return found
}

// AddUnmarshal 注册解析函数
func (mgr *Manager) AddUnmarshal(m UnmarshalFunc, ext ...string) error {
	for _, e := range ext {
		if e == "" || e == "." {
			return errExtNotAllowEmpty
		}

		if e[0] != '.' {
			e = "." + e
		}

		e = strings.ToLower(e)
		if mgr.Exists(e) {
			return fmt.Errorf("已经存在该扩展名 %s 对应的解析函数", ext)
		}
		mgr.unmarshals[e] = m
	}

	return nil
}

// SetUnmarshal 修改指定扩展名关联的解析函数，不存在则添加。
func (mgr *Manager) SetUnmarshal(m UnmarshalFunc, ext ...string) error {
	for _, e := range ext {
		if e == "" || e == "." {
			return errExtNotAllowEmpty
		}

		if e[0] != '.' {
			e = "." + e
		}

		mgr.unmarshals[strings.ToLower(e)] = m
	}

	return nil
}

// File 获取文件路径，相对于当前配置目录
func (mgr *Manager) File(path string) string {
	return filepath.Join(mgr.dir, path)
}

// LoadFile 加载指定的配置文件内容到 v 中
func (mgr *Manager) LoadFile(path string, v interface{}) error {
	r, err := os.Open(mgr.File(path))
	if err != nil {
		return err
	}
	defer r.Close()

	return mgr.Load(r, filepath.Ext(path), v)
}

// Load 加载指定的配置文件内容到 v 中
func (mgr *Manager) Load(r io.Reader, typ string, v interface{}) error {
	typ = strings.ToLower(typ)
	unmarshal, found := mgr.unmarshals[typ]
	if !found {
		return fmt.Errorf("无效的配置文件类型：%s", typ)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if err = unmarshal(data, v); err != nil {
		return err
	}

	if s, ok := v.(Sanitizer); ok {
		return s.Sanitize()
	}
	return nil
}
