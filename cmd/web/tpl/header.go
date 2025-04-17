// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package tpl

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/issue9/web"
)

type fileHeaderBuildFunc = func([]string) string

var headerBuilders = map[string]fileHeaderBuildFunc{
	".go": singleCStyle,

	".ts": singleCStyle,
	".js": singleCStyle,

	".rs": singleCStyle,

	".cs": singleCStyle,
	".css": multiCStyle,

	".c":   singleCStyle,
	".cxx": singleCStyle,
	".cc":  singleCStyle,
	".cpp": singleCStyle,
	".h":   singleCStyle,
	".hpp": singleCStyle,
	".hxx": singleCStyle,
	".m":   singleCStyle,

	".java": singleCStyle,
	".kt":   singleCStyle,
	".kts":  singleCStyle,

	".swift": singleCStyle,

	".py": singlePythonStyle,

	".clj": func(s []string) string { return singleStyle([]byte("; "), s) },

	".sh":   singlePythonStyle,
	".rb":   singlePythonStyle,
	".ps1":  singlePythonStyle,
	".psm1": singlePythonStyle,

	".yaml": singlePythonStyle,
	".yml":  singlePythonStyle,
	".toml": singlePythonStyle,
}

// 为指定的扩展名的文件插入文件头
func insertFileHeaders(dir, header, extsStr string) error {
	exts := strings.Split(extsStr, ",")
	for index, ext := range exts {
		if ext[0] != '.' {
			exts[index] = "." + ext
		}
	}

	headers := strings.Split(strings.TrimSpace(header), "\n")

	for _, ext := range exts {
		if _, found := headerBuilders[ext]; !found {
			return web.NewLocaleError("unsupported file extension '%s'", ext)
		}
	}

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if slices.Index(exts, ext) < 0 {
			return nil
		}

		build, found := headerBuilders[ext]
		if !found {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		f, err := os.OpenFile(path, os.O_WRONLY, modePerm)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write([]byte(build(headers))); err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	})
}

func singleCStyle(s []string) string { return singleStyle([]byte("// "), s) }

func singlePythonStyle(s []string) string { return singleStyle([]byte("# "), s) }

func multiCStyle(s []string) string {
	return multiStyle([]byte("/*"), []byte(" */"), []byte(" * "), s)
}

func singleStyle(prefix []byte, s []string) string {
	l := 5 * len(s)
	for _, v := range s {
		l += len(v)
	}

	b := make([]byte, 0, l)
	for _, v := range s {
		b = append(b, prefix...)
		b = append(b, v...)
		b = append(b, '\n')
	}
	b = append(b, '\n') // 空行

	return string(b)
}

func multiStyle(start, end, prefix []byte, s []string) string {
	l := +5*len(s) + len(start) + len(end)
	for _, v := range s {
		l += len(v)
	}

	b := make([]byte, 0, l)
	b = append(b, start...)
	b = append(b, '\n')
	for _, v := range s {
		b = append(b, prefix...)
		b = append(b, v...)
		b = append(b, '\n')
	}
	b = append(b, end...)
	b = append(b, '\n') // 换行
	b = append(b, '\n') // 空行

	return string(b)
}
