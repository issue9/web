// SPDX-License-Identifier: MIT

// Package parser 文档内容分析
package parser

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
	"github.com/issue9/web/cmd/web/internal/restdoc/pkg"
	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

var errSyntax = web.NewLocaleError("syntax error")

// Parser 文档分析对象
type Parser struct {
	pkgsM  sync.Mutex
	pkgs   []*pkg.Package
	search schema.SearchFunc
	fset   *token.FileSet

	media []string

	// api 的部分功能是依赖 restdoc 的，
	// 在 restdoc 未解析的情况下，所有的 api 注释都要缓存。
	apiComments []*comments

	parsed bool
	l      *logger.Logger
}

type comments struct {
	lines   []string
	pos     token.Pos
	modPath string
}

// New 声明 RESTDoc 对象
func New(l *logger.Logger) *Parser {
	doc := &Parser{
		pkgs: make([]*pkg.Package, 0, 10),
		fset: token.NewFileSet(),

		apiComments: make([]*comments, 0, 100),

		l: l,
	}

	doc.search = func(s string) *pkg.Package {
		if p, found := sliceutil.At(doc.pkgs, func(pkg *pkg.Package, _ int) bool { return pkg.Path == s }); found {
			return p
		}
		return nil
	}

	return doc
}

// AddDir 添加 root 下的内容
//
// 仅在调用 [RESTDoc.Openapi3] 之前添加有效果。
// root 添加的目录；
func (p *Parser) AddDir(ctx context.Context, root string, recursive bool) {
	if p.parsed {
		panic("已经解析完成，无法再次添加！")
	}
	pkg.ScanDir(ctx, p.fset, root, recursive, p.append, p.l)
}

// line 返回 pos 的行号
func (p *Parser) line(pos token.Pos) int { return p.fset.Position(pos).Line }

func (p *Parser) file(pos token.Pos) string { return p.fset.File(pos).Name() }

func (p *Parser) append(pp *pkg.Package) {
	p.pkgsM.Lock()
	defer p.pkgsM.Unlock()

	if sliceutil.Exists(p.pkgs, func(pkg *pkg.Package, _ int) bool { return pkg.Path == pp.Path }) {
		p.l.Error(web.Phrase("package %s with the same name.", pp.Path), "", 0)
		return
	}

	p.pkgs = append(p.pkgs, pp)
}

// SaveAs 保存为 yaml 或 json 文件
//
// 根据后缀名名确定保存的文件类型，目前仅支持 json 和 yaml。
func (p *Parser) SaveAs(ctx context.Context, path string, tag ...string) error {
	var m func(any) ([]byte, error)
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		m = yaml.Marshal
	case ".json":
		m = func(v any) ([]byte, error) {
			return json.MarshalIndent(v, "", "\t")
		}
	default:
		return web.NewLocaleError("only support yaml and json")
	}

	data, err := m(p.OpenAPI(ctx, tag...))
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, os.ModePerm)
}

// OpenAPI 转换成 openapi3.T 对象
//
// tags 如果非空，则表示仅返回带这些标签的 API。
//
// NOTE: 已经执行了 [openapi3.T.Validate]。
func (p *Parser) OpenAPI(ctx context.Context, tags ...string) *openapi3.T {
	p.parsed = true // 阻止 doc.AddDir

	t := schema.NewOpenAPI("3.0.0")
	wg := &sync.WaitGroup{}

	for _, pp := range p.pkgs {
		select {
		case <-ctx.Done():
			p.l.Warning(pkg.Cancelled)
			return nil
		default:
			wg.Add(1)
			go func(pp *pkg.Package) {
				defer wg.Done()
				p.parsePackage(ctx, t, pp, tags)
			}(pp)
		}
	}
	wg.Wait()

	// 下面的操作依赖上面的完成，所以需要两个 wg 变量

	wg = &sync.WaitGroup{}
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
						p.parseAPI(t, c.modPath, suffix, c.lines[index+1:], p.line(c.pos)+index, p.file(c.pos), tags)
					}
				}
			}(c)
		}
	}
	wg.Wait()

	// BUG(caixw) 无法验证 #/components/responses 下的引用?
	/*
		if err := t.Validate(ctx); err != nil {
			p.l.Error(err, "", 0)
			return nil
		}
	*/

	return t
}

func (p *Parser) parsePackage(ctx context.Context, t *openapi3.T, pack *pkg.Package, tags []string) {
	wg := &sync.WaitGroup{}
	for _, f := range pack.Files {
		select {
		case <-ctx.Done():
			p.l.Warning(pkg.Cancelled)
			return
		default:
			wg.Add(1)
			go func(f *ast.File) {
				defer wg.Done()
				p.parseFile(t, pack.Path, f, tags)
			}(f)
		}
	}
	wg.Wait()
}

func (p *Parser) parseFile(t *openapi3.T, importPath string, f *ast.File, tags []string) {
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
					if t.Info != nil {
						p.parseAPI(t, importPath, suffix, lines[index+1:], p.line(c.Pos()), p.file(c.Pos()), tags)
					} else {
						p.apiComments = append(p.apiComments, &comments{
							lines:   lines, // 保存所有行，而不是从当前页开始，方便后续判断正确的行号。
							pos:     c.Pos(),
							modPath: importPath,
						})
					}
				case "restdoc":
					p.parseRESTDoc(t, importPath, suffix, lines[index+1:], p.line(c.Pos())+index, p.file(c.Pos()), tags)
				}
				continue LOOP
			}
		}
	}
}

// 只要 tag 有一个元素在 enableTags 或是 enableTags 为空都返回 false
func isIgnoreTag(enableTags []string, tag ...string) bool {
	if len(enableTags) == 0 {
		return false
	}

	for _, t := range tag {
		if sliceutil.Exists(enableTags, func(tt string, _ int) bool { return tt == t }) {
			return false
		}
	}
	return true
}