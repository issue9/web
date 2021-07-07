// SPDX-License-Identifier: MIT

package content

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestContext_ServeFileFS(t *testing.T) {
	a := assert.New(t)
	fsys := os.DirFS("./")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := &Context{
		Response: w,
		Request:  r,
	}

	// p = context.go
	w.Body.Reset()
	data, err := fs.ReadFile(fsys, "context.go")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "context.go", "", nil))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data)

	// index = context.go
	w.Body.Reset()
	data, err = fs.ReadFile(fsys, "context.go")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "", "context.go", nil))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data)

	// p=gob, index=gob.go
	w.Body.Reset()
	data, err = fs.ReadFile(fsys, "gob/gob.go")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "gob", "gob.go", nil))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data)

		// p=gob, index=gob.go, headers != nil
	w.Body.Reset()
	data, err = fs.ReadFile(fsys, "gob/gob.go")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "gob", "gob.go", map[string]string{"Test": "ttt"}))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data).
		Equal(w.Header().Get("Test"), "ttt")

	w.Body.Reset()
	a.ErrorIs(ctx.ServeFS(fsys, "gob", "", nil), os.ErrNotExist).
		Empty(w.Body.Bytes())

	w.Body.Reset()
	a.ErrorIs(ctx.ServeFS(fsys, "gob", "not-exists.go", nil), os.ErrNotExist).
		Empty(w.Body.Bytes())

	w.Body.Reset()
	a.ErrorIs(ctx.ServeFS(fsys, "not-exists", "file.go", nil), os.ErrNotExist).
		Empty(w.Body.Bytes())
}
