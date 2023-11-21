module github.com/issue9/web/cmd/web

go 1.21

require (
	github.com/caixw/gobuild v1.7.2
	github.com/getkin/kin-openapi v0.120.0
	github.com/issue9/assert/v3 v3.1.0
	github.com/issue9/cmdopt v0.13.0
	github.com/issue9/localeutil v0.24.0
	github.com/issue9/query/v3 v3.1.2
	github.com/issue9/source v0.7.0
	github.com/issue9/term/v3 v3.2.4
	github.com/issue9/version v1.0.7
	github.com/issue9/web v0.85.0
	golang.org/x/mod v0.14.0
	golang.org/x/text v0.14.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/issue9/web => ../../

require (
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/invopop/yaml v0.2.0 // indirect
	github.com/issue9/config v0.6.1 // indirect
	github.com/issue9/conv v1.3.4 // indirect
	github.com/issue9/errwrap v0.3.1 // indirect
	github.com/issue9/logs/v7 v7.1.0 // indirect
	github.com/issue9/mux/v7 v7.3.3 // indirect
	github.com/issue9/scheduled v0.15.1 // indirect
	github.com/issue9/sliceutil v0.15.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
)
