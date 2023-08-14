// SPDX-License-Identifier: MIT

package schema

import (
	"context"
	"go/token"
	"slices"
	"sync"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/internal/restdoc/openapi"
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

		if i := slices.IndexFunc(pkgs, func(pkg *pkg.Package) bool { return pkg.Path == s }); i >= 0 {
			return pkgs[i]
		}
		return nil
	}
}

func TestSearchFunc_New(t *testing.T) {
	a := assert.New(t, false)
	f := buildSearchFunc(a)
	modPath := "github.com/issue9/web/cmd/web/internal/restdoc/schema/testdata"
	modRef := refReplacer.Replace(modPath)

	// #components/schemas/abc
	t.Run(refPrefix, func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		refPath := refPrefix + ".admin.notFound"
		ref, err := f.New(tt, modPath, refPath, false)
		a.NotError(err).Equal(ref.Ref, refPath).Nil(ref.Value)
	})

	// NotFound
	t.Run("NotFound", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		refPath := modPath + "/admin.notFound"
		ref, err := f.New(tt, modPath, refPath, false)
		a.Error(err, web.NewLocaleError("not found %s", refPath)).Nil(ref)
	})

	// Generic
	t.Run("Generic", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "Generic", false)
		a.ErrorString(err, "unsupported generics type").Nil(ref)
	})

	// Generic IndexExpr
	t.Run("泛型 IndexExpr", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "IntGeneric", false)
		a.NotError(err).NotNil(ref)
		v, found := ref.Value.Properties["Type"]
		a.True(found).NotNil(v)
	})

	// Generic Generics[int,Admin]
	t.Run("泛型 Generics[int,Admin]", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath+"/admin", modPath+".Generics[int, Admin]", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Ref, refPrefix+modRef+".Generics-int--Admin-")

		v, found := ref.Value.Properties["F1"]
		a.True(found).NotNil(v)

		v, found = ref.Value.Properties["F2"]
		a.True(found).NotNil(v).
			Equal(v.Value.AllOf[0].Ref, refPrefix+modRef+".admin.Admin")

		v, found = ref.Value.Properties["P"]
		a.True(found).NotNil(v).
			Equal(v.Value.Type, openapi3.TypeInteger)
	})

	// Generic IndexListExpr
	t.Run("泛型 IndexListExpr", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, modPath+"/admin.IntUserGenerics", false)
		a.NotError(err).NotNil(ref)

		v, found := ref.Value.Properties["F1"]
		a.True(found).NotNil(v).
			Equal(v.Ref, refPrefix+modRef+".Generic-int-")

		v, found = ref.Value.Properties["F2"]
		a.True(found).NotNil(v).
			Equal(v.Ref, refPrefix+modRef+".User")
	})

	t.Run("Sexes", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "Sexes", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, refPrefix+modRef+".Sex")
	})

	t.Run("Int64UserGenerics", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, modPath+"/admin.Int64UserGenerics", false)
		a.NotError(err).NotNil(ref)

		v, found := ref.Value.Properties["G1"]
		a.True(found).NotNil(v).
			Equal(v.Ref, refPrefix+modRef+".Generics-int64--User-")
	})

	t.Run("[]bool", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "[]bool", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Value.Type, openapi3.TypeBoolean)
	})

	// time.Time
	t.Run("time.Time", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "time.Time", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Format, "date-time")
	})

	// time.Duration
	t.Run("time.Duration", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "time.Duration", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeString)
	})

	// 枚举数组
	t.Run("[]Sex", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "[]Sex", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, refPrefix+modRef+".Sex")

		sex, found := tt.GetSchema(modRef + ".Sex")
		a.True(found).NotNil(sex).
			Equal(sex.Value.Description, "Sex 表示性别\n@enum female male unknown\n@type string\n").
			Equal(sex.Value.Type, "string").
			Equal(sex.Value.Enum, []string{"female", "male", "unknown"})
	})

	// 对象数组
	t.Run("[]User", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, "[]User", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, refPrefix+modRef+".User")
		u, found := tt.GetSchema(modRef + ".User")
		a.True(found).NotNil(u).
			Empty(u.Value.Description). // 单行 doc，赋值给了 title
			Equal(u.Value.Title, "用户信息 doc").
			Equal(u.Value.Type, openapi3.TypeObject)

		name := u.Value.Properties["Name"]
		a.Equal(name.Value.AllOf[0].Ref, refPrefix+modRef+".String").
			Equal(name.Value.Title, "姓名")

		sex := u.Value.Properties["sex"]
		a.Equal(sex.Value.AllOf[0].Ref, refPrefix+modRef+".Sex").
			True(sex.Value.XML.Attribute).
			Equal(sex.Value.Title, "性别")

		age := u.Value.Properties["age"]
		a.Empty(age.Ref).
			Equal(age.Value.Title, "年龄").
			Equal(age.Value.Type, openapi3.TypeInteger)

		st := u.Value.Properties["struct"]
		a.Empty(st.Ref).
			Equal(st.Value.Title, "struct doc").
			Equal(st.Value.Type, openapi3.TypeObject)
	})

	// XMLName
	t.Run("XMLName", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, modPath+"/admin.Admin", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.XML.Name, "admin")
	})

	// admin.User
	t.Run("admin.User", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, modPath+"/admin.User", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Ref, refPrefix+modRef+".admin.User")
		u, found := tt.GetSchema(modRef + ".admin.User")
		a.True(found).NotNil(u).
			Equal(u.Value.Title, "User testdata.User").
			Equal(u.Value.AllOf[0].Ref, refPrefix+modRef+".User")
	})

	// admin.Admin
	t.Run("admin.Admin", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(tt, modPath, modPath+"/admin.Admin", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeObject)

		admin, found := tt.GetSchema(modRef + ".admin.Admin")
		a.True(found).NotNil(admin)
		name := admin.Value.Properties["Name"]
		a.Equal(name.Value.AllOf[0].Ref, refPrefix+modRef+".String").
			Equal(name.Value.Title, "姓名")

		u1 := admin.Value.Properties["U1"]
		a.Empty(u1.Ref).
			Equal(u1.Value.Title, "u1").
			Equal(u1.Value.Type, openapi3.TypeArray).
			Equal(u1.Value.Items.Ref, refPrefix+modRef+".User")

		u2 := admin.Value.Properties["u2"]
		a.Empty(u2.Ref).
			Equal(u2.Value.Title, "u2").
			True(u2.Value.Nullable).
			Equal(u2.Value.AllOf[0].Ref, refPrefix+modRef+".User")

		u3 := admin.Value.Properties["u3"]
		a.Nil(u3)

		u4 := admin.Value.Properties["U4"]
		a.Equal(u4.Ref, refPrefix+modRef+".admin.User")
	})
}
