// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/mimetypetest"
)

func init() {
	chk := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	chk(message.SetString(language.Chinese, "test", "中文"))
	chk(message.SetString(language.SimplifiedChinese, "test", "简体"))
	chk(message.SetString(language.TraditionalChinese, "test", "繁体"))
	chk(message.SetString(language.English, "test", "english"))
}

func newContext(a *assert.Assertion,
	w http.ResponseWriter,
	r *http.Request,
	outputCharset encoding.Encoding,
	InputCharset encoding.Encoding) *Context {
	return &Context{
		server: newServer(a),

		Response:       w,
		Request:        r,
		OutputCharset:  outputCharset,
		OutputMimetype: mimetypetest.TextMarshal,

		InputCharset:  InputCharset,
		InputMimetype: mimetypetest.TextUnmarshal,
	}
}

func TestContext_Vars(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newServer(a).newContext(w, r)

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

func TestServer_newContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	srv := newServer(a)
	logwriter := new(bytes.Buffer)
	srv.Logs().ERROR().SetOutput(logwriter)

	// 错误的 accept
	logwriter.Reset()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	a.Panic(func() {
		srv.newContext(w, r)
	})
	a.True(logwriter.Len() > 0)

	// 错误的 accept-charset
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("Accept-Charset", "unknown")
	a.Panic(func() {
		srv.newContext(w, r)
	})
	a.True(logwriter.Len() > 0)

	// 错误的 content-type,无输入内容
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Content-Type", ";charset=utf-8")
	a.Panic(func() {
		srv.newContext(w, r)
	})

	// 错误的 content-type,有输入内容
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("[]"))
	r.Header.Set("Content-Type", ";charset=utf-8")
	a.Panic(func() {
		srv.newContext(w, r)
	})
	a.True(logwriter.Len() > 0)

	// 错误的 content-type，且有输入内容
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("content-type", buildContentType(mimetypetest.Mimetype, "utf-"))
	a.Panic(func() {
		srv.newContext(w, r)
	})

	// 错误的 Accept-Language
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx")
	ctx := srv.newContext(w, r)
	a.NotNil(ctx)
	a.Equal(ctx.OutputTag, language.MustParse("zh-hans"))
	a.Equal(ctx.Server(), srv)

	// 正常，指定 Accept-Language
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7")
	ctx = srv.newContext(w, r)
	a.NotNil(ctx)
	a.Equal(logwriter.Len(), 0).
		Equal(ctx.InputCharset, nil).
		Equal(ctx.OutputMimetypeName, mimetype.DefaultMimetype).
		Equal(ctx.OutputTag, language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	a.NotPanic(func() {
		ctx = srv.newContext(w, r)
	})
	a.NotNil(ctx).
		Equal(logwriter.Len(), 0).
		Equal(ctx.InputCharset, nil).
		Equal(ctx.OutputMimetypeName, mimetype.DefaultMimetype)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("content-type", buildContentType(mimetypetest.Mimetype, "utf-8"))
	a.NotPanic(func() {
		ctx = srv.newContext(w, r)
	})
	a.NotNil(ctx).
		Equal(logwriter.Len(), 0).
		True(charsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, mimetype.DefaultMimetype)
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newContext(a, w, r, nil, nil)

	// 未缓存
	a.Nil(ctx.body)
	data, err := ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 读取缓存内容
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用 Nop 即 utf-8 编码
	w.Body.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	ctx = newContext(a, w, r, encoding.Nop, encoding.Nop)
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用不同的编码
	w.Body.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkdata1))
	r.Header.Set("Accept", "*/*")
	ctx = newContext(a, w, r, encoding.Nop, simplifiedchinese.GB18030)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkstr1)
	a.Equal(ctx.body, data)
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("test,123"))
	w := httptest.NewRecorder()
	ctx := newContext(a, w, r, nil, nil)

	obj := &mimetypetest.TextObject{}
	a.True(ctx.Read(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	o := &struct{}{}
	a.False(ctx.Read(o))
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(a, w, r, nil, nil)
	obj := &mimetypetest.TextObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, map[string]string{"contEnt-type": "json", "content-lanGuage": "zh-hans"}))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
	a.Equal(w.Header().Get("content-type"), "json")
	a.Equal(w.Header().Get("content-language"), "zh-hans")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(a, w, r, encoding.Nop, encoding.Nop)
	obj = &mimetypetest.TextObject{Name: "test", Age: 1234}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,1234")
	a.Equal(w.Header().Get("content-language"), "") // 未指定

	// 输出 nil
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(a, w, r, encoding.Nop, encoding.Nop)
	ctx.OutputTag = language.MustParse("zh-Hans")
	a.NotError(ctx.Marshal(http.StatusCreated, nil, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "")
	a.Equal(w.Header().Get("content-language"), "zh-Hans") // 指定了输出语言

	// 输出 Nil
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(a, w, r, encoding.Nop, encoding.Nop)
	a.NotError(ctx.Marshal(http.StatusCreated, mimetype.Nil, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), mimetypetest.Nil)

	// 输出不同编码的内容
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(a, w, r, simplifiedchinese.GB18030, encoding.Nop)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkstr2, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), gbkdata2)
}

func TestContext_Render(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(a, w, r, nil, nil)
	obj := &mimetypetest.TextObject{Name: "test", Age: 123}
	ctx.Render(http.StatusCreated, obj, nil)
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(a, w, r, nil, nil)
	obj1 := &struct{ Name string }{Name: "name"}
	ctx.Render(http.StatusCreated, obj1, nil)
	a.Equal(w.Code, http.StatusInternalServerError)
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx := newContext(a, w, r, nil, nil)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	// httptest.NewRequest 会直接将  remote-addr 赋值为 192.0.2.1 无法测试
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("x-real-ip", "192.168.1.1:8080")
	ctx = newContext(a, w, r, nil, nil)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	ctx = newContext(a, w, r, nil, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newContext(a, w, r, nil, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Remote-Addr", "192.168.2.0")
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newContext(a, w, r, nil, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}

func TestContext_Created(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx := newContext(a, w, r, nil, nil)
	ctx.Created(&mimetypetest.TextObject{Name: "test", Age: 123}, "")
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w.Body.Reset()
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx = newContext(a, w, r, nil, nil)
	ctx.Created(&mimetypetest.TextObject{Name: "test", Age: 123}, "/test")
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`).
		Equal(w.Header().Get("Location"), "/test")
}
