module github.com/issue9/web/cmd/web

go 1.27.0

require (
	github.com/caixw/gobuild v1.9.1
	github.com/goccy/go-yaml v1.19.2
	github.com/issue9/assert/v6 v6.0.0
	github.com/issue9/cmdopt v0.16.0
	github.com/issue9/errwrap v0.4.1
	github.com/issue9/localeutil v0.35.0
	github.com/issue9/logs/v7 v7.7.0
	github.com/issue9/messages v0.1.0
	github.com/issue9/source v0.13.1
	github.com/issue9/web v0.104.4
	github.com/otiai10/copy v1.14.1
	golang.org/x/mod v0.42.0
	golang.org/x/text v0.43.0
)

replace github.com/issue9/web => ../../

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/andybalholm/brotli v1.2.4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/issue9/cache v0.21.0 // indirect
	github.com/issue9/config v0.10.0 // indirect
	github.com/issue9/conv v1.4.0 // indirect
	github.com/issue9/mux/v10 v10.0.0-beta.3 // indirect
	github.com/issue9/query/v3 v3.2.0 // indirect
	github.com/issue9/scheduled v0.24.0 // indirect
	github.com/issue9/sliceutil v0.19.0 // indirect
	github.com/issue9/term/v3 v3.5.0 // indirect
	github.com/jellydator/ttlcache/v3 v3.4.1 // indirect
	github.com/klauspost/compress v1.20.0 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
)
