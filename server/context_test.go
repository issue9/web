// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/content"
	"github.com/issue9/web/content/text"
)

func TestContext_Vars(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newServer(a).NewContext(w, r)

	type (
		t1 int
		t2 int64
		t3 = t2
	)
	var (
		v1 t1 = 1
		v2 t2 = 1
		v3 t3 = 1
	)

	ctx.Vars[v1] = 1
	ctx.Vars[v2] = 2
	ctx.Vars[v3] = 3

	a.Equal(ctx.Vars[v1], 1).Equal(ctx.Vars[v2], 3)
}

func TestServer_NewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	srv := newServer(a)
	logwriter := new(bytes.Buffer)
	srv.Logs().DEBUG().SetOutput(logwriter)

	// 错误的 accept
	logwriter.Reset()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	a.Panic(func() {
		srv.NewContext(w, r)
	})
	a.True(logwriter.Len() > 0)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", content.DefaultMimetype)
	a.NotPanic(func() {
		ctx := srv.NewContext(w, r)
		a.NotNil(ctx).
			Equal(logwriter.Len(), 0).
			True(content.CharsetIsNop(ctx.InputCharset)).
			Equal(ctx.OutputMimetypeName, content.DefaultMimetype)
	})
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("test,123"))
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", text.Mimetype+"; charset=utf-8")
	ctx := newServer(a).NewContext(w, r)

	obj := &text.TestObject{}
	a.True(ctx.Read(obj, 41110))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	// 触发 ctx.Error 退出
	a.PanicString(func() {
		o := &struct{}{}
		ctx.Read(o, 41110)
	}, "0") // 简单地抛出 0，让 recovery 捕获处理。
}

func TestContext_Render(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", text.Mimetype)
	r.Header.Set("Accept", text.Mimetype)
	ctx := newServer(a).NewContext(w, r)
	obj := &text.TestObject{Name: "test", Age: 123}
	ctx.Render(http.StatusCreated, obj, nil)
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")

	// 触发 ctx.Error 退出
	a.PanicString(func() {
		r = httptest.NewRequest(http.MethodPost, "/path", nil)
		w = httptest.NewRecorder()
		r.Header.Set("Content-Type", text.Mimetype)
		r.Header.Set("Accept", text.Mimetype)
		ctx = newServer(a).NewContext(w, r)
		obj1 := &struct{ Name string }{Name: "name"}
		ctx.Render(http.StatusCreated, obj1, nil)
	}, "0") // 简单地抛出 0，让 recovery 捕获处理。
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx := newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	// httptest.NewRequest 会直接将  remote-addr 赋值为 192.0.2.1 无法测试
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("x-real-ip", "192.168.1.1:8080")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Remote-Addr", "192.168.2.0")
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}

func TestContext_Created(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	ctx := newServer(a).NewContext(w, r)
	ctx.Created(&text.TestObject{Name: "test", Age: 123}, "")
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w.Body.Reset()
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	ctx = newServer(a).NewContext(w, r)
	ctx.Created(&text.TestObject{Name: "test", Age: 123}, "/test")
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`).
		Equal(w.Header().Get("Location"), "/test")
}
