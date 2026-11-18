// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/issue9/mux/v10/header"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/problems"
	"github.com/issue9/web/selector"
)

// Client 用于访问由 [Server] 提供服务的客户端
//
// NOTE: 远程如果不是 [Server] 实现的服务，可能无法正确处理返回对象。
type Client struct {
	client   *http.Client
	nego     *Negotiation
	selector selector.Selector

	mediaType MediaType

	requestIDKey string
	requestIDGen func() string
}

// NewClient 创建 [Client] 实例
//
// client 如果为空，表示采用 &http.Client{} 作为默认值；
// marshal 指定提交数据的媒体类型，如果对应的 [MediaType.Marshal] 需要 ctx 参数，可能 panic；
// reqIDKey 表示 x-request-id 的报头名称，如果为空则表示不需要；
// reqIDGen 表示生成 x-request-id 值的方法；
func NewClient(
	client *http.Client, nego *Negotiation, s selector.Selector, marshal string, reqIDKey string, reqIDGen func() string,
) *Client {
	if client == nil {
		client = &http.Client{}
	}

	if reqIDKey != "" && reqIDGen == nil {
		panic("当前 requestIDKey 不为空时 requestIDGen 也不能为空")
	}

	mt, found := sliceutil.At(nego.mediaTypes, func(mt MediaType, _ int) bool { return mt.Name() == marshal })
	if !found {
		panic("参数 marshal 未找到")
	}

	return &Client{
		client:   client,
		nego:     nego,
		selector: s,

		mediaType: mt,

		requestIDKey: reqIDKey,
		requestIDGen: reqIDGen,
	}
}

func (c *Client) Get[R any, E any](path string) (*R, *Problem, error) {
	return c.Do[R, E](http.MethodGet, path, nil)
}

func (c *Client) Delete[R any, E any](path string) (*R, *Problem, error) {
	return c.Do[R, E](http.MethodDelete, path, nil)
}

func (c *Client) Post[R any, E any](path string, body any) (*R, *Problem, error) {
	return c.Do[R, E](http.MethodPost, path, body)
}

func (c *Client) Put[R any, E any](path string, body any) (*R, *Problem, error) {
	return c.Do[R, E](http.MethodPut, path, body)
}

func (c *Client) Patch[R any, E any](path string, body any) (*R, *Problem, error) {
	return c.Do[R, E](http.MethodPatch, path, body)
}

// Do 开始新的请求
//
// body 为提交的对象，最终是由初始化参数的 marshal 进行编码；
func (c *Client) Do[R any, E any](method, path string, body any) (*R, *Problem, error) {
	req, err := c.NewRequest(method, path, body)
	if err != nil {
		return nil, nil, err
	}
	rsp, err := c.Client().Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer rsp.Body.Close() // ignore error

	return c.ParseResponse[R, E](rsp)
}

// ParseResponse 从 [http.Response] 解析并获取返回对象
//
// R 表示正常解析之后返回的类型，该值不能为指针类型；
// E 表示出错时返回的 Problem.Extensions 的类型，大部分情况该值为 any 即可；
func (c *Client) ParseResponse[R any, E any](rsp *http.Response) (r *R, p *Problem, err error) {
	if rsp.ContentLength == 0 { // 204 可能为空
		return nil, nil, nil
	}

	var reader io.Reader = rsp.Body
	encName := rsp.Header.Get(header.ContentEncoding)
	encodingReader, err := c.nego.contentEncoding(encName, reader)
	if err != nil {
		return nil, nil, err
	}
	reader = encodingReader

	defer func() {
		// rsp.Body 由传入方负责关闭。
		// 需要注意后续代码对于 err 的赋值方式，是否是覆盖。
		err = errors.Join(err, encodingReader.Close())
	}()

	var unmarshal UnmarshalFunc
	var inputCharset encoding.Encoding
	if h := rsp.Header.Get(header.ContentType); h != "" {
		if unmarshal, inputCharset, err = c.nego.contentType(h); err != nil {
			return
		}

		if unmarshal == nil {
			err = NewError("not found unmarshaler for the server content-type %s", h)
			return
		}
	} else {
		err = NewError("the server miss content-type header")
		return
	}

	if !qheader.CharsetIsNop(inputCharset) {
		reader = transform.NewReader(reader, inputCharset.NewDecoder())
	}

	if problems.IsProblem(rsp.StatusCode) {
		var e E
		p = &Problem{} // TODO
		p.Extensions = e
		if err = unmarshal(reader, p); err != nil {
			return
		}
		return nil, p, nil
	}

	var resp R
	if err = unmarshal(reader, &resp); err != nil {
		return
	}
	return &resp, nil, nil
}

// NewRequest 生成 [http.Request]
//
// body 为需要提交的对象；
func (c *Client) NewRequest(method, path string, body any) (resp *http.Request, err error) {
	var data []byte
	if body != nil {
		if data, err = c.mediaType.Marshal(nil, body); err != nil {
			return nil, err
		}
	}

	if path, err = c.URL(path); err != nil {
		return nil, err
	}

	bodyIsEmpty := false
	var r *http.Request
	if len(data) == 0 {
		bodyIsEmpty = true
		r, err = http.NewRequest(method, path, nil)
	} else {
		r, err = http.NewRequest(method, path, bytes.NewBuffer(data))
	}
	if err != nil {
		return nil, err
	}

	if !bodyIsEmpty {
		r.Header.Set(header.ContentType, qheader.BuildContentType(c.mediaType.Name(), header.UTF8))
		r.Header.Set(header.ContentLanguage, c.nego.Language().String())
		// 提交的内容始终不压缩，所以就不提交 content-encoding 报头了。
	}

	r.Header.Set(header.Accept, c.nego.clientAcceptHeader)
	r.Header.Set(header.AcceptEncoding, c.nego.acceptEncodingHeader)
	r.Header.Set(header.AcceptLanguage, c.nego.acceptLanguageHeader)
	if c.requestIDKey != "" {
		r.Header.Set(c.requestIDKey, c.requestIDGen())
	}

	return r, nil
}

// URL 生成一条访问地址
func (c *Client) URL(path string) (string, error) {
	u, err := c.selector.Next()
	if err != nil {
		return "", err
	}
	return u + path, nil
}

func (c *Client) Client() *http.Client { return c.client }

// NewClient 基于当前对象的相关字段创建 [Client] 对象
//
// 功能与 [NewClient] 相同，缺少的参数直接采用 [Server] 关联的字段。
func (s *Server) NewClient(c *http.Client, sel selector.Selector, m string) *Client {
	return NewClient(c, s.Negotiation(), sel, m, s.requestIDKey, s.UniqueID)
}
