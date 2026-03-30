module github.com/issue9/web/cmd/web

go 1.26.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/caixw/gobuild v1.8.8
	github.com/goccy/go-yaml v1.19.2
	github.com/issue9/assert/v4 v4.3.1
	github.com/issue9/cmdopt v0.13.1
	github.com/issue9/errwrap v0.3.3
	github.com/issue9/localeutil v0.33.0
	github.com/issue9/logs/v7 v7.6.10
	github.com/issue9/source v0.12.8
	github.com/issue9/web v0.104.4
	github.com/otiai10/copy v1.14.1
	golang.org/x/mod v0.34.0
	golang.org/x/text v0.35.0
)

replace github.com/issue9/web => ../../

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/issue9/cache v0.19.6 // indirect
	github.com/issue9/config v0.9.5 // indirect
	github.com/issue9/conv v1.3.7 // indirect
	github.com/issue9/mux/v9 v9.2.3 // indirect
	github.com/issue9/query/v3 v3.1.4 // indirect
	github.com/issue9/scheduled v0.22.5 // indirect
	github.com/issue9/sliceutil v0.17.0 // indirect
	github.com/issue9/term/v3 v3.4.5 // indirect
	github.com/jellydator/ttlcache/v3 v3.4.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
)
