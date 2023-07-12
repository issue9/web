// SPDX-License-Identifier: MIT

// Package parser 文档内容分析
package parser

import (
	"context"
	"go/ast"
	"go/token"
	"strings"
	"sync"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
	"github.com/issue9/web/cmd/web/internal/restdoc/pkg"
	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

var errSyntax = localeutil.Error("syntax error")

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
		if p, found := sliceutil.At(doc.pkgs, func(pkg *pkg.Package) bool { return pkg.Path == s }); found {
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
func (doc *Parser) AddDir(ctx context.Context, root string, recursive bool) {
	if doc.parsed {
		panic("已经解析完成，无法再次添加！")
	}
	pkg.ScanDir(ctx, doc.fset, root, recursive, doc.append, doc.l)
}

// line 返回 p 的行号
func (doc *Parser) line(p token.Pos) int { return doc.fset.Position(p).Line }

func (doc *Parser) file(p token.Pos) string { return doc.fset.File(p).Name() }

func (doc *Parser) append(p *pkg.Package) {
	doc.pkgsM.Lock()
	defer doc.pkgsM.Unlock()

	if sliceutil.Exists(doc.pkgs, func(pkg *pkg.Package) bool { return pkg.Path == p.Path }) {
		doc.l.Log(logger.Unknown, localeutil.Phrase("package %s with the same name.", p.Path), "", 0)
		return
	}

	doc.pkgs = append(doc.pkgs, p)
}

// OpenAPI 转换成 openapi3.T 对象
func (doc *Parser) OpenAPI(ctx context.Context) *openapi3.T {
	doc.parsed = true // 阻止 doc.AddDir

	t := schema.NewOpenAPI()

	wg := &sync.WaitGroup{}
	for _, p := range doc.pkgs {
		select {
		case <-ctx.Done():
			doc.l.LogWithoutPos(logger.Cancelled, context.Canceled)
			return nil
		default:
			wg.Add(1)
			go func(p *pkg.Package) {
				defer wg.Done()
				doc.parsePackage(ctx, t, p)
			}(p)
		}
	}
	wg.Wait()

	for _, c := range doc.apiComments {
		for index, line := range c.lines {
			if len(line) <= 2 {
				continue
			}
			if tag, suffix := utils.CutTag(line[2:]); suffix != "" && strings.ToLower(tag) == "api" {
				doc.parseAPI(t, c.modPath, suffix, c.lines[index+1:], doc.line(c.pos)+index, doc.file(c.pos))
			}
		}
	}

	return t
}

func (doc *Parser) parsePackage(ctx context.Context, t *openapi3.T, pkg *pkg.Package) {
	wg := &sync.WaitGroup{}
	for _, f := range pkg.Files {
		select {
		case <-ctx.Done():
			doc.l.LogWithoutPos(logger.Cancelled, context.Canceled)
			return
		default:
			wg.Add(1)
			go func(f *ast.File) {
				defer wg.Done()
				doc.parseFile(t, pkg.Path, f)
			}(f)
		}
	}
	wg.Wait()
}

func (doc *Parser) parseFile(t *openapi3.T, importPath string, f *ast.File) {
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
						doc.parseAPI(t, importPath, suffix, lines[index+1:], doc.line(c.Pos()), doc.file(c.Pos()))
					} else {
						doc.apiComments = append(doc.apiComments, &comments{
							lines:   lines, // 保存所有行，而不是从当前页开始，方便后续判断正确的行号。
							pos:     c.Pos(),
							modPath: importPath,
						})
					}
				case "restdoc":
					doc.parseRESTDoc(t, importPath, suffix, lines[index+1:], doc.line(c.Pos())+index, doc.file(c.Pos()))
				}
				continue LOOP
			}
		}
	}
}
