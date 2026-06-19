module github.com/conall/outalator

go 1.25.0

require (
	github.com/coreos/go-oidc/v3 v3.9.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/sessions v1.2.2
	github.com/lib/pq v1.10.9
	golang.org/x/oauth2 v0.27.0
	google.golang.org/grpc v1.56.3
	google.golang.org/protobuf v1.33.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	// modernc.org/sqlite and its transitive deps are required by -tags sqlite.
	// Use `GOFLAGS="-tags=sqlite" go mod tidy` (or `make tidy-sqlite`) to keep
	// them in sync; a plain `go mod tidy` without the tag will remove them.
	// All entries remain // indirect because go mod tidy cannot see the import
	// in factory_sqlite.go without the build tag — this is a known Go limitation.
	modernc.org/libc v1.72.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.52.0
)
