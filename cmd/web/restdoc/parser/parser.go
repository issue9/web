// SPDX-License-Identifier: MIT

// Package parser 文档内容分析
package parser

import (
	"context"
	"go/ast"
	"go/token"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/web"
	"golang.org/x/tools/go/packages"

	"github.com/issue9/web/cmd/web/restdoc/logger"
	"github.com/issue9/web/cmd/web/restdoc/openapi"
	"github.com/issue9/web/cmd/web/restdoc/pkg"
	"github.com/issue9/web/cmd/web/restdoc/schema"
	"github.com/issue9/web/cmd/web/restdoc/utils"
)

// Parser 文档分析对象
type Parser struct {
	schema *schema.Schema

	media   []string                      // 全局可用 media type
	resps   map[string]*openapi3.Response // 全局可用 response
	headers []pair
	cookies []pair

	// api 的部分功能是依赖 restdoc 的，
	// 在 restdoc 未解析的情况下，所有的 api 注释都要缓存。
	apiComments  []*comments
	apiCommentsM sync.Mutex

	parsed bool
	l      *logger.Logger

	prefix string
	tags   []string
}

type pair struct {
	key  string
	desc string
}

type comments struct {
	lines   []string
	pos     token.Pos
	modPath string
}

// New 声明 [Parser] 对象
//
// prefix 为所有的 API 地址加上统一的前缀；
// tags 如果非空，则表示仅返回带这些标签的 API；
func New(l *logger.Logger, prefix string, tags []string) *Parser {
	return &Parser{
		schema: schema.New(l),

		apiComments: make([]*comments, 0, 100),

		l: l,

		prefix: prefix,
		tags:   tags,
	}
}

// AddDir 添加 root 下的内容
//
// 仅在调用 [Parser.Parse] 之前添加有效果。
// root 添加的目录；
func (p *Parser) AddDir(ctx context.Context, root string, recursive bool) {
	if p.parsed {
		panic("已经解析完成，无法再次添加！")
	}
	p.schema.Packages().ScanDir(ctx, root, recursive)
}

// line 返回 pos 的行号
func (p *Parser) line(pos token.Pos) int { return p.schema.Packages().FileSet().Position(pos).Line }

func (p *Parser) file(pos token.Pos) string {
	return p.schema.Packages().FileSet().Position(pos).Filename
}

// Parse 解析由 [Parser.AddDir] 加载的内容
func (p *Parser) Parse(ctx context.Context) *openapi.OpenAPI {
	p.parsed = true // 阻止 p.AddDir

	t := openapi.New("3.0.0")
	wg := &sync.WaitGroup{}

	p.schema.Packages().Range(func(pp *packages.Package) bool {
		select {
		case <-ctx.Done():
			p.l.Warning(pkg.Cancelled)
			return false
		default:
			wg.Add(1)
			go func(pp *packages.Package) {
				defer wg.Done()
				p.parsePackage(ctx, t, pp)
			}(pp)
		}
		return true
	})

	wg.Wait()

	// 下面的操作依赖上面的完成，所以需要两个 wg 变量

	wg = &sync.WaitGroup{}
	p.apiCommentsM.Lock()
	defer p.apiCommentsM.Unlock()

	for _, c := range p.apiComments {
		select {
		case <-ctx.Done():
			p.l.Warning(pkg.Cancelled)
			return nil
		default:
			wg.Add(1)
			go func(c *comments) {
				defer wg.Done()

				for index, line := range c.lines {
					if len(line) <= 2 {
						continue
					}
					if tag, suffix := utils.CutTag(line[2:]); suffix != "" && strings.ToLower(tag) == "api" {
						p.parseAPI(ctx, t, c.modPath, suffix, c.lines[index+1:], p.line(c.pos)+index, p.file(c.pos))
					}
				}
			}(c)
		}
	}
	wg.Wait()

	if err := t.Doc().Validate(ctx); err != nil {
		p.l.Error(err, "", 0)
		return nil
	}

	return t
}

func (p *Parser) parsePackage(ctx context.Context, t *openapi.OpenAPI, pack *packages.Package) {
	wg := &sync.WaitGroup{}
	for _, f := range pack.Syntax {
		select {
		case <-ctx.Done():
			p.l.Warning(pkg.Cancelled)
			return
		default:
			wg.Add(1)
			go func(f *ast.File) {
				defer wg.Done()
				p.parseFile(ctx, t, pack.PkgPath, f)
			}(f)
		}
	}
	wg.Wait()
}

func (p *Parser) parseFile(ctx context.Context, t *openapi.OpenAPI, importPath string, f *ast.File) {
LOOP:
	for _, c := range f.Comments {
		lines := strings.Split(c.Text(), "\n")
		if len(lines) <= 2 { // 少于 2 行的肯定不是
			continue
		}

		for index, line := range lines {
			if len(line) < 6 || // 最起码得包含 # api 5 个字符
				line[0] != '#' || !unicode.IsSpace(rune(line[1])) { // # 之后至少一个空格
				continue
			}

			if tag, suffix := utils.CutTag(line[2:]); suffix != "" {
				switch strings.ToLower(tag) {
				case "api":
					if t.Doc().Info != nil {
						p.parseAPI(ctx, t, importPath, suffix, lines[index+1:], p.line(c.Pos()), p.file(c.Pos()))
					} else {
						p.apiCommentsM.Lock()
						p.apiComments = append(p.apiComments, &comments{
							lines:   lines, // 保存所有行，而不是从当前页开始，方便后续判断正确的行号。
							pos:     c.Pos(),
							modPath: importPath,
						})
						p.apiCommentsM.Unlock()
					}
				case "restdoc":
					p.parseRESTDoc(ctx, t, importPath, suffix, lines[index+1:], p.line(c.Pos())+index, p.file(c.Pos()))
				}
				continue LOOP
			}
		}
	}
}

// 只要 tag 有一个元素在 p.tags 或是 p.tags 为空都返回 false
func (p *Parser) isIgnoreTag(tag ...string) bool {
	if len(p.tags) == 0 {
		return false
	}

	for _, t := range tag {
		if slices.IndexFunc(p.tags, func(tt string) bool { return tt == t }) >= 0 {
			return false
		}
	}
	return true
}

// tag 标签名称，比如 @version；
// size 标签最少需要的参数；
// filename 和 ln 表示出错的位置，分别为文件名和行号；
func (p *Parser) syntaxError(tag string, size int, filename string, ln int) {
	p.l.Error(web.NewLocaleError("%s requires at least %d parameters", tag, size), filename, ln)
}

// 将 p 和 name 合并为一个类型名称
//
// p 表示包地址，name 为类型名称，如果 name 为泛型，则泛型中的名称也会加 p 作为包地址。
func buildPath(p, name string) string {
	if strings.HasPrefix(name, openapi.ComponentSchemaPrefix) {
		return name
	}

	var g bool // 是否包含泛型参数
	last := strings.LastIndexByte(name, '[')
	if last > 0 && name[len(name)-1] == ']' {
		g = true
		list := strings.Split(name[last+1:len(name)-1], ",")
		for i, t := range list {
			if ii := strings.IndexByte(t, '.'); ii < 0 {
				t = p + "." + t
			}
			list[i] = t
		}

		strs := strings.Join(list, ",")
		name = name[:last] + "[" + strs + "]"
	}

	if !g {
		last = len(name)
	}
	if strings.IndexByte(name[:last], '.') > 0 {
		return name
	}

	var index int // []* 等修饰符与类型之间的位置
	var flag bool // 是否在 [] 之间
LOOP:
	for i := range len(name) {
		index = i
		switch c := name[i]; {
		case c == '[':
			if flag { // 不支持 [[
				index = 0
				break LOOP
			}
			flag = true
		case c == ']':
			if !flag { // 不支持 ]]
				index = 0
				break LOOP
			}
			flag = false
		case c == '*':
			if flag { // 不支持 [*
				index = 0
				break LOOP
			}
		case c > '0' && c < '9':
			if !flag {
				index = 0
				break LOOP
			}
		default:
			if flag {
				index = 0
			}
			break LOOP
		}
	}

	pun := strings.TrimSpace(name[:index]) // 类型前的符号部分
	name = strings.TrimSpace(name[index:])

	return pun + p + "." + name
}
