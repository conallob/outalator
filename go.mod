module github.com/conall/outalator

go 1.27.0

require (
	github.com/coreos/go-oidc/v3 v3.21.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/sessions v1.4.0
	github.com/lib/pq v1.12.3
	golang.org/x/oauth2 v0.37.0
	google.golang.org/grpc v1.85.0
	google.golang.org/protobuf v1.36.12
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.61.1
)

require (
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.8 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.1.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261026165753-2ffab32bce7e // indirect
	// modernc.org/sqlite and its transitive deps are required by -tags sqlite.
	// Use `GOFLAGS="-tags=sqlite" go mod tidy` (or `make tidy-sqlite`) to keep
	// them in sync; a plain `go mod tidy` without the tag will remove them.
	// All entries remain // indirect because go mod tidy cannot see the import
	// in factory_sqlite.go without the build tag — this is a known Go limitation.
	modernc.org/libc v1.80.0 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
)
