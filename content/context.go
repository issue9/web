// SPDX-License-Identifier: MIT

package content

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"

	"github.com/issue9/qheader"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/serialization"
)

// DefaultCharset 默认的字符集
const DefaultCharset = "utf-8"

// 需要作比较，所以得是经过 http.CanonicalHeaderKey 处理的标准名称。
var (
	contentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	contentLanguageKey = http.CanonicalHeaderKey("Content-Language")
)

// Context 单次请求生成的上下文数据
//
// NOTE: 用户不应该直接引用该对象，而是 server.Context。
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request

	// 指定输出时所使用的媒体类型，以及名称
	OutputMimetype     serialization.MarshalFunc
	OutputMimetypeName string

	// 输出到客户端的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	OutputCharset     encoding.Encoding
	OutputCharsetName string

	// 客户端内容所使用的媒体类型
	InputMimetype serialization.UnmarshalFunc

	// 客户端内容所使用的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	InputCharset encoding.Encoding

	// 输出语言的相关设置项
	OutputTag     language.Tag
	LocalePrinter *message.Printer

	// 保存着从 http.Request.Body 中获取的内容。
	//
	// body 用于缓存从 http.Request.Body 中读取的内容；
	// read 表示是否需要从 http.Request.Body 读取内容。
	body []byte
	read bool
}

// NewContext 从用户请求中构建一个 Context 实例
//
// 如果不合规则，会以指定的状码返回并向 l 输出信息。
func (c *Content) NewContext(l *log.Logger, w http.ResponseWriter, r *http.Request) (*Context, int) {
	printLog := func(format string, v ...interface{}) {
		if l != nil {
			l.Printf(format, v...)
		}
	}

	header := r.Header.Get("Accept")
	outputMimetypeName, marshal, found := c.Mimetypes().MarshalFunc(header)
	if !found {
		printLog("未找到符合报头 %s 的解码函数", header)
		return nil, http.StatusNotAcceptable
	}

	header = r.Header.Get("Accept-Charset")
	outputCharsetName, outputCharset := acceptCharset(header)
	if outputCharsetName == "" {
		printLog("未找到符合报头 %s 的字符集", header)
		return nil, http.StatusNotAcceptable
	}

	header = r.Header.Get(contentTypeKey)
	inputMimetype, inputCharset, err := c.conentType(header)
	if err != nil {
		printLog(err.Error())
		return nil, http.StatusUnsupportedMediaType
	}

	tag := c.acceptLanguage(r.Header.Get("Accept-Language"))

	ctx := &Context{
		Response:           w,
		Request:            r,
		OutputMimetype:     marshal,
		OutputMimetypeName: outputMimetypeName,
		OutputCharset:      outputCharset,
		OutputCharsetName:  outputCharsetName,
		OutputTag:          tag,
		LocalePrinter:      c.newLocalePrinter(tag),
		InputMimetype:      inputMimetype,
		InputCharset:       inputCharset,
	}

	return ctx, 0
}

// Body 获取用户提交的内容
//
// 相对于 ctx.Request.Body，此函数可多次读取。不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.read {
		return ctx.body, nil
	}

	if ctx.body, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		return nil, err
	}
	ctx.read = true

	if charsetIsNop(ctx.InputCharset) {
		return ctx.body, nil
	}

	d := ctx.InputCharset.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(ctx.body), d)
	ctx.body, err = ioutil.ReadAll(reader)
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象
func (ctx *Context) Unmarshal(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	if ctx.InputMimetype != nil {
		return ctx.InputMimetype(body, v)
	}
	return nil
}

// Marshal 将 v 解码并发送给客户端
//
// status 表示输出的状态码，如果出错，可能输出的不是该状态码；
// v 输出的对象，若是一个 nil 值，则不会向客户端输出任何内容；
// 若是需要正常输出一个 nil 类型到客户端（比如JSON 中的 null），
// 可以传递一个 *struct{} 值，或是自定义实现相应的解码函数；
// headers 报头信息，如果已经存在于 ctx.Response 将覆盖 ctx.Response 中的值，
// 如果需要指定一个特定的 Content-Type 和 Content-Language，
// 可以在 headers 中指定，否则使用当前的编码和语言名称；
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	header := ctx.Response.Header()
	var contentTypeFound, contentLanguageFound bool
	for k, v := range headers {
		k = http.CanonicalHeaderKey(k)

		contentTypeFound = contentTypeFound || k == contentTypeKey
		contentLanguageFound = contentLanguageFound || k == contentLanguageKey
		header.Set(k, v)
	}

	if !contentTypeFound {
		ct := buildContentType(ctx.OutputMimetypeName, ctx.OutputCharsetName)
		header.Set(contentTypeKey, ct)
	}

	if !contentLanguageFound && ctx.OutputTag != language.Und {
		header.Set(contentLanguageKey, ctx.OutputTag.String())
	}

	if v == nil {
		ctx.Response.WriteHeader(status)
		return nil
	}

	// 没有指定编码函数，NewContext 阶段是允许 OutputMimetype 为空的，所以只能在此处判断。
	if ctx.OutputMimetype == nil {
		ctx.Response.WriteHeader(http.StatusNotAcceptable)
		return nil
	}
	data, err := ctx.OutputMimetype(v)
	if err != nil {
		return err
	}

	// 注意 WriteHeader 调用顺序。
	// https://github.com/golang/go/issues/17083
	//
	// NOTE: 此处由原来的 errorhandler.WriteHeader 改为 ctx.Response.WriteHeader
	// 即 Marshal 函数也接受 errorhandler 的捕获，不作特殊处理。
	ctx.Response.WriteHeader(status)

	if charsetIsNop(ctx.OutputCharset) {
		_, err = ctx.Response.Write(data)
		return err
	}

	w := transform.NewWriter(ctx.Response, ctx.OutputCharset.NewEncoder())
	defer func() {
		if err2 := w.Close(); err2 != nil {
			if err != nil {
				err2 = fmt.Errorf("在处理错误 %w 时再次抛出错误 %s", err, err2.Error())
			}
			err = err2
		}
	}()
	_, err = w.Write(data)
	return err
}

// acceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func acceptCharset(header string) (name string, enc encoding.Encoding) {
	if header == "" || header == "*" {
		return DefaultCharset, nil
	}

	var err error
	accepts := qheader.Parse(header, "*")
	for _, apt := range accepts {
		enc, err = htmlindex.Get(apt.Value)
		if err != nil { // err != nil 表示未找到，继续查找
			continue
		}

		// 转换成官方的名称
		name, err = htmlindex.Name(enc)
		if err != nil {
			name = apt.Value // 不存在，直接使用用户上传的名称
		}

		return name, enc
	}

	return "", nil
}

func (c *Content) acceptLanguage(header string) language.Tag {
	if header == "" {
		return c.Tag()
	}

	al := qheader.Parse(header, "*")
	tags := make([]language.Tag, 0, len(al))
	for _, l := range al {
		tags = append(tags, language.Make(l.Value))
	}

	tag, _, _ := c.locale.Builder().Matcher().Match(tags...)
	return tag
}

// conentType 从 content-type 报头解析出需要用到的解码函数
func (c *Content) conentType(header string) (serialization.UnmarshalFunc, encoding.Encoding, error) {
	var (
		mt      = DefaultMimetype
		charset = DefaultCharset
	)

	if header != "" {
		mts, params, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, nil, err
		}
		mt = mts
		if charset = params["charset"]; charset == "" {
			charset = DefaultCharset
		}
	}

	f, found := c.Mimetypes().UnmarshalFunc(mt)
	if !found {
		return nil, nil, fmt.Errorf("未注册的解码函数 %s", mt)
	}

	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}
