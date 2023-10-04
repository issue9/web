// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/compress"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/mimetypes"
)

// Client 用于访问远程的客户端
type Client struct {
	url        string
	client     *http.Client
	compresses *compress.Compresses

	mts         *mtsType
	marshal     func(any) ([]byte, error)
	marshalName string
}

// NewClient 创建 Client 实例
//
// url 远程服务的地址基地址，url 不能以 / 结尾。比如 https://example.com:8080/s1；
// marshal 对输入数据的编码方式；
// mt 所有返回数据可用的解码方式；
func NewClient(url, marshalName string, mt []*Mimetype, compresses []*Compress) *Client {
	mts := mimetypes.New[BuildMarshalFunc, UnmarshalFunc](len(mt))
	for _, m := range mt {
		mts.Add(m.Type, m.MarshalBuilder, m.Unmarshal, m.ProblemType)
	}

	c := compress.NewCompresses(len(compresses), false)
	for i, e := range compresses {
		if err := e.sanitize(); err != nil {
			panic(err.AddFieldParent("compresses[" + strconv.Itoa(i) + "]"))
		}
		c.Add(e.Name, e.Compressor, e.Types...)
	}

	return newClient(url, marshalName, mts, c)
}

// NewClient 采用 [Server] 的编码和压缩方式创建 Client 对象
//
// 参数可参考 [NewClient]。
func (srv *Server) NewClient(url, marshalName string) *Client {
	return newClient(url, marshalName, srv.mimetypes, srv.compresses)
}

func newClient(url, marshalName string, m *mtsType, c *compress.Compresses) *Client {
	if l := len(url); l > 0 && url[l-1] == '/' {
		url = url[:l-1]
	}

	var marshal MarshalFunc
	if mm := m.Search(marshalName); mm != nil {
		marshal = mm.MarshalBuilder(nil)
	}
	if marshal == nil {
		panic(fmt.Sprintf("未找到 %s 指定的编码方法", marshalName))
	}

	return &Client{
		url:        url,
		client:     &http.Client{},
		compresses: c,

		mts:         m,
		marshal:     marshal,
		marshalName: marshalName,
	}
}

func (c *Client) Get(path string, resp any, problem *RFC7807) error {
	return c.Do(http.MethodGet, path, nil, resp, problem)
}

func (c *Client) Delete(path string, resp any, problem *RFC7807) error {
	return c.Do(http.MethodDelete, path, nil, resp, problem)
}

func (c *Client) Post(path string, req, resp any, problem *RFC7807) error {
	return c.Do(http.MethodPost, path, req, resp, problem)
}

func (c *Client) Put(path string, req, resp any, problem *RFC7807) error {
	return c.Do(http.MethodPut, path, req, resp, problem)
}

func (c *Client) Patch(path string, req, resp any, problem *RFC7807) error {
	return c.Do(http.MethodPatch, path, req, resp, problem)
}

// Do 开始新的请求
//
// req 为提交的对象，最终是由初始化参数的 marshal 进行编码；
// resp 为返回的数据的写入对象，必须是指针类型；
// problem 为返回出错时的写入对象；
// 非 HTTP 状态码错误返回 err；
func (c *Client) Do(method, path string, req, resp any, problem *RFC7807) (err error) {
	r, err := c.NewRequest(method, path, req)
	if err != nil {
		return err
	}
	rsp, err := c.client.Do(r)
	if err != nil {
		return err
	}

	return c.ParseResponse(rsp, resp, problem)
}

// ParseResponse 从 [http.Response] 解析并获取返回对象
//
// 如果正常状态，将内容解码至 resp，如果出错了，则解码至 problem。
// 其它情况下返回错误信息。
func (c *Client) ParseResponse(rsp *http.Response, resp any, problem *RFC7807) (err error) {
	var size int
	if h := rsp.Header.Get("Content-Length"); h != "" {
		if h == "0" {
			return nil
		}

		size, err = strconv.Atoi(h)
		if err != nil {
			return err
		}
	}
	if size == 0 {
		return nil
	}

	var reader io.Reader = rsp.Body
	encName := rsp.Header.Get(header.ContentEncoding)
	reader, err = c.compresses.ContentEncoding(encName, reader)
	if err != nil {
		return nil
	}

	var inputMimetype UnmarshalFunc
	var inputCharset encoding.Encoding
	if h := rsp.Header.Get(header.ContentType); h != "" {
		if inputMimetype, inputCharset, err = c.mts.ContentType(h); err != nil {
			return err
		}

		if inputMimetype == nil {
			return NewLocaleError("not found unmarshaler for the server content-type %s", h)
		}
	} else {
		return NewLocaleError("the server miss content-type header")
	}

	if !header.CharsetIsNop(inputCharset) {
		reader = transform.NewReader(reader, inputCharset.NewDecoder())
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		return inputMimetype(data, problem)
	}
	return inputMimetype(data, resp)
}

// NewRequest 生成 [http.Request]
//
// body 为需要提交的对象；
func (c *Client) NewRequest(method, path string, body any) (resp *http.Request, err error) {
	var data []byte
	if body != nil {
		data, err = c.marshal(body)
		if err != nil {
			return nil, err
		}
	}

	var r *http.Request
	if len(data) == 0 {
		r, err = http.NewRequest(method, c.URL(path), nil)
	} else {
		r, err = http.NewRequest(method, c.URL(path), bytes.NewBuffer(data))
	}
	if err != nil {
		return nil, err
	}
	r.Header.Set(header.ContentType, header.BuildContentType(c.marshalName, header.UTF8Name))
	r.Header.Set(header.Accept, c.mts.AcceptHeader())
	r.Header.Set(header.AcceptEncoding, c.compresses.AcceptEncodingHeader())

	return r, nil
}

// URL 生成一条访问地址
func (c *Client) URL(path string) string { return c.url + path }
