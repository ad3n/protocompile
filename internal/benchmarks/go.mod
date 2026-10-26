module github.com/ad3n/protocompile/internal/benchmarks

go 1.27.3

require (
	github.com/ad3n/protocompile v0.14.1
	github.com/jhump/protoreflect v1.14.1 // MUST NOT be updated to v1.15 or higher
	github.com/stretchr/testify v1.11.1
	golang.org/x/sync v0.22.0
	google.golang.org/protobuf v1.37.0-rc.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/petermattis/goid v0.0.0-20260716134002-a9b348f0a2b9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/tidwall/btree v1.8.1 // indirect
	golang.org/x/exp v0.0.0-20260709172345-9ea1abe57597 // indirect
	golang.org/x/net v0.57.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ad3n/protocompile => ../..
