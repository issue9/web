// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package create 用于创建新项目的子命令
package create

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/issue9/utils"
	"github.com/issue9/term/prompt"
	"github.com/issue9/term/colors"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/webconfig"
)

// Usage 当前子命令的用法
func Usage(output io.Writer) {
	fmt.Fprintln(output, `构建一个新的 web 项目

语法：web create [mod]
mod 为一个可选参数，如果指定了，则会直接使用此值作为模块名，
若不指定，则会通过之后的交互要求用户指定。模块名中的最后一
路径名称，会作为目录名称创建于当前目录下。`)
}

// Do 执行子命令
func Do(output io.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ask := prompt.New('\n',os.Stdin, output,colors.Green)

	if len(os.Args) == 3 {
		return createMod(os.Args[2], wd, ask)
	}

	mod, err := ask.String("模块名", "")
	if err != nil {
		return err
	}
	if mod == "" {
		return errors.New("模块名不能为空")
	}
	return createMod(mod, wd, ask)
}

// 创建包的目录结构。
//
// mod 是模块名称，比如 github.com/issue9/web；
// wd 表示当前的工作目录，项目会在此目录中创建。
func createMod(mod, wd string, ask *prompt.Prompt) error {
	name := filepath.Base(mod)
	path, err := filepath.Abs(filepath.Join(wd, name))
	if err != nil {
		return err
	}

	// 创建文件夹
	if utils.FileExists(path) {
		cover, err := ask.Bool("存在同名文件夹，是否覆盖", false)
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
	content := fmt.Sprintf(gomod, mod, web.Version)
	if err = dumpFile(filepath.Join(path, "go.mod"), content); err != nil {
		return err
	}

	if err = createModules(path); err != nil {
		return err
	}

	return createCmd(path, "cmd/main", mod)
}

// 创建 cmd 目录内容
//
// path 表示项目的根目录；
// dir 表示相对于 path 的 cmd 目录名称；
// mod 表示模块的包名。
func createCmd(path, dir, mod string) error {
	path = filepath.Join(path, dir)

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	// 输出 main.go
	data := fmt.Sprintf(maingo, mod+"/modules")
	if err := dumpFile(filepath.Join(path, "main.go"), data); err != nil {
		return err
	}

	return createConfig(path, "appconfig")
}

// 创建配置文件目录，并输出默认的配置内容。
//
// path 为项目的根目录；
// dir 为配置文件的目录名称，相对于 path 目录。
func createConfig(path, dir string) error {
	path = filepath.Join(path, dir)

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	// 输出 logs.xml
	if err := dumpFile(filepath.Join(path, web.DefaultLogsFilename), logs); err != nil {
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
	return dumpFile(filepath.Join(path, web.DefaultConfigFilename), string(data))
}

// 创建模块目录，并输出默认的配置内容。
//
// path 为项目的根目录
func createModules(path string) error {
	path = filepath.Join(path, "modules")

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	// 输出 modules.go
	return dumpFile(filepath.Join(path, "modules.go"), modulesgo)
}

func dumpFile(path string, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()

	_, err = file.WriteString(content)
	return err
}
