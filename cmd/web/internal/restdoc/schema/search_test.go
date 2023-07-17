// SPDX-License-Identifier: MIT

package schema

import (
	"context"
	"go/token"
	"sync"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/internal/restdoc/pkg"
)

func buildSearchFunc(a *assert.Assertion) SearchFunc {
	ctx := context.Background()
	fset := token.NewFileSet()
	l := loggertest.New(a)

	var pkgs []*pkg.Package
	var pkgsM sync.Mutex
	af := func(p *pkg.Package) {
		pkgsM.Lock()
		defer pkgsM.Unlock()
		pkgs = append(pkgs, p)
	}
	pkg.ScanDir(ctx, fset, "./testdata", true, af, l.Logger)

	return func(s string) *pkg.Package {
		pkgsM.Lock()
		defer pkgsM.Unlock()

		r, found := sliceutil.At(pkgs, func(pkg *pkg.Package) bool { return pkg.Path == s })
		if found {
			return r
		}
		return nil
	}
}

func TestSearchFunc_NewSchema(t *testing.T) {
	a := assert.New(t, false)
	f := buildSearchFunc(a)
	modPath := "github.com/issue9/web/cmd/web/internal/restdoc/schema/testdata"

	t.Run("[]bool", func(t *testing.T) {
		a := assert.New(t, false)
		tt := NewOpenAPI()

		ref, err := f.New(tt, modPath, "[]bool", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Value.Type, openapi3.TypeBoolean)
	})

	// 枚举数组
	t.Run("[]Sex", func(t *testing.T) {
		a := assert.New(t, false)
		tt := NewOpenAPI()

		ref, err := f.New(tt, modPath, "[]Sex", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, modPath+".Sex")

		sex := tt.Components.Schemas[modPath+".Sex"]
		a.NotNil(sex).
			Equal(sex.Value.Description, "Sex 表示性别\n@enum female male unknown\n").
			Equal(sex.Value.Type, openapi3.TypeInteger).
			Equal(sex.Value.Enum, []string{"female", "male", "unknown"})
	})

	// 对象数组
	t.Run("[]User", func(t *testing.T) {
		a := assert.New(t, false)
		tt := NewOpenAPI()

		ref, err := f.New(tt, modPath, "[]User", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, modPath+".User")
		u := tt.Components.Schemas[modPath+".User"]
		a.NotNil(u).
			Equal(u.Value.Description, "用户信息 doc\n").
			Equal(u.Value.Type, openapi3.TypeObject)

		name := u.Value.Properties["Name"]
		a.Equal(name.Value.AllOf[0].Ref, modPath+".String").
			Equal(name.Value.Description, "姓名\n")

		sex := u.Value.Properties["sex"]
		a.Equal(sex.Value.AllOf[0].Ref, modPath+".Sex").
			True(sex.Value.XML.Attribute).
			Equal(sex.Value.Description, "性别\n")

		age := u.Value.Properties["age"]
		a.Empty(age.Ref).
			Equal(age.Value.Description, "年龄\n").
			Equal(age.Value.AllOf[0].Value.Type, openapi3.TypeInteger)
	})

	// admin.User
	t.Run("admin.User", func(t *testing.T) {
		a := assert.New(t, false)
		tt := NewOpenAPI()

		ref, err := f.New(tt, modPath, modPath+"/admin.User", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Ref, modPath+".User")
	})

	// admin.Admin
	t.Run("admin.Admin", func(t *testing.T) {
		a := assert.New(t, false)
		tt := NewOpenAPI()

		ref, err := f.New(tt, modPath, modPath+"/admin.Admin", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeObject)

		admin := tt.Components.Schemas[modPath+"/admin.Admin"]
		name := admin.Value.Properties["Name"]
		a.Equal(name.Value.AllOf[0].Ref, modPath+".String").
			Equal(name.Value.Description, "姓名\n")

		u1 := admin.Value.Properties["U1"]
		a.Empty(u1.Ref).
			Equal(u1.Value.Description, "u1\n").
			Equal(u1.Value.AllOf[0].Value.Type, openapi3.TypeArray).
			Equal(u1.Value.AllOf[0].Value.Items.Ref, modPath+".User")

		u2 := admin.Value.Properties["u2"]
		a.Empty(u2.Ref).
			Equal(u2.Value.Description, "u2\n").
			True(u2.Value.Nullable).
			Equal(u2.Value.AllOf[0].Ref, modPath+".User")

		u3 := admin.Value.Properties["u3"]
		a.Nil(u3)

		u4 := admin.Value.Properties["U4"]
		a.Equal(u4.Ref, modPath+".User")
	})
}

func TestWrap(t *testing.T) {
	a := assert.New(t, false)

	ref := openapi3.NewSchemaRef("ref", openapi3.NewBoolSchema())
	ref2 := wrap(ref, "", nil, false)
	a.Equal(ref2, ref)

	ref2 = wrap(ref, "123", nil, false)
	a.NotEqual(ref2, ref).
		Equal(ref2.Value.AllOf[0].Value, ref.Value).
		Equal(ref2.Value.Description, "123")

	ref2 = wrap(ref, "123", nil, true)
	a.NotEqual(ref2, ref).
		Equal(ref2.Value.AllOf[0].Value, ref.Value).
		Equal(ref2.Value.Description, "123").
		True(ref2.Value.Nullable)
}
