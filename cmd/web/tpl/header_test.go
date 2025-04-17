// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package tpl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	xcopy "github.com/otiai10/copy"

	"github.com/issue9/web"
)

func TestSingleCStyle(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(singleCStyle([]string{}), "\n")
	a.Equal(singleCStyle([]string{""}), "// \n\n")

	a.Equal(singleCStyle([]string{
		"SPDX-FileCopyrightText: 2025 caixw",
		"SPDX-License-Identifier: MIT",
	}), `// SPDX-FileCopyrightText: 2025 caixw
// SPDX-License-Identifier: MIT

`)
}

func TestMultipCStyle(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(multiCStyle([]string{}), "/*\n */\n\n")
	a.Equal(multiCStyle([]string{""}), "/*\n * \n */\n\n")

	a.Equal(multiCStyle([]string{
		"SPDX-FileCopyrightText: 2025 caixw",
		"SPDX-License-Identifier: MIT",
	}), `/*
 * SPDX-FileCopyrightText: 2025 caixw
 * SPDX-License-Identifier: MIT
 */

`)
}

func TestInsertFileHeader(t *testing.T) {
	a := assert.New(t, false)
	dir := "./testdata/out/fileheaders"
	headers := "SPDX-FileCopyrightText: 2025 caixw\nSPDX-License-Identifier: MIT\n"
	wantCHeaders := "// SPDX-FileCopyrightText: 2025 caixw\n// SPDX-License-Identifier: MIT\n"
	wantPyHeaders := "# SPDX-FileCopyrightText: 2025 caixw\n# SPDX-License-Identifier: MIT\n"

	a.NotError(xcopy.Copy("./testdata/template", dir))
	a.Equal(insertFileHeaders(dir, headers, ".not-exists"), web.NewLocaleError("unsupported file extension '%s'", ".not-exists"))
	a.NotError(os.RemoveAll(dir))

	// .go

	a.NotError(xcopy.Copy("./testdata/template", dir))
	a.NotError(insertFileHeaders(dir, headers, ".go"))

	c := getFileContent(a, filepath.Join(dir, "template.go"))
	a.True(strings.HasPrefix(c, wantCHeaders), "%s", c)

	c = getFileContent(a, filepath.Join(dir, "sub/sub_test.go"))
	a.True(strings.HasPrefix(c, wantCHeaders), "%s", c)

	c = getFileContent(a, filepath.Join(dir, "abc.cpp")) // 未指定 cpp
	a.False(strings.HasPrefix(c, wantCHeaders), "%s", c)

	a.NotError(os.RemoveAll(dir))

	// .go .cpp .py

	a.NotError(xcopy.Copy("./testdata/template", dir))
	a.NotError(insertFileHeaders(dir, headers, ".go,.cpp,py"))

	c = getFileContent(a, filepath.Join(dir, "template.go"))
	a.True(strings.HasPrefix(c, wantCHeaders), "%s", c)

	c = getFileContent(a, filepath.Join(dir, "abc.cpp"))
	a.True(strings.HasPrefix(c, wantCHeaders), "%s", c)

	c = getFileContent(a, filepath.Join(dir, "abc.py"))
	a.True(strings.HasPrefix(c, wantPyHeaders), "%s", c)

	a.NotError(os.RemoveAll(dir))
}

func getFileContent(a *assert.Assertion, path string) string {
	data, err := os.ReadFile(path)
	a.NotError(err)
	return string(data)
}
