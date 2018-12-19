// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package create 用于创建新项目的子命令
package create

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/issue9/utils"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/web"
	"github.com/issue9/web/app"
	"github.com/issue9/web/internal/webconfig"
)

// Do 执行子命令
func Do(output *os.File) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ask := newAsker(os.Stdin, output)

	if len(os.Args) == 3 {
		return createMod(os.Args[2], wd, ask)
	}

	mod, err := ask.Ask("模块名", "")
	if err != nil {
		return err
	}
	if mod == "" {
		return errors.New("模块名不能为空")
	}
	return createMod(mod, wd, ask)
}

// 创建包的目录结构。
func createMod(mod, wd string, ask *asker) error {
	name := filepath.Base(mod)
	path, err := filepath.Abs(filepath.Join(wd, name))
	if err != nil {
		return err
	}

	// 创建文件夹
	if utils.FileExists(path) {
		cover, err := ask.AskBool("存在同名文件夹，是否覆盖", false)
		if err != nil {
			return err
		}
		if cover {
			goto MOD
		}
	} else {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return err
		}
	}

MOD:
	content := fmt.Sprintf(`module %s

required github.com/issue9/web v%s`, mod, web.Version)
	err = dumpFile(filepath.Join(path, "go.mod"), []byte(content))
	if err != nil {
		return err
	}

	if err = createConfig(path, "appconfig"); err != nil {
		return err
	}

	return createModules(path)
}

// 创建配置文件目录，并输出默认的配置内容。
// path 为项目的根目录
// dir 为配置文件的目录名称
func createConfig(path, dir string) error {
	path = filepath.Join(path, dir)

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	// 输出 logs.xml
	if err := dumpFile(filepath.Join(path, app.LogsFilename), logs); err != nil {
		return err
	}

	// web.yaml
	conf := &webconfig.WebConfig{
		HTTPS:  false,
		Domain: "localhost",
		Port:   8080,
	}
	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return dumpFile(filepath.Join(path, app.ConfigFilename), data)
}

// 创建模块目录，并输出默认的配置内容。
// path 为项目的根目录
func createModules(path string) error {
	path = filepath.Join(path, "modules")

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	// TODO 输出 modules.go

	return nil
}

func dumpFile(path string, content []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()

	_, err = file.Write(content)
	return err
}
