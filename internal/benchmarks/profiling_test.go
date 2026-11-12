package benchmarks

import (
	"strings"
	"testing"

	"github.com/ad3n/protocompile"
	"github.com/ad3n/protocompile/internal/testing/googleapis"
	"github.com/ad3n/protocompile/parser/fastscan"
)

func BenchmarkGoogleapisMemoryCompile(b *testing.B) {
	benchmarkMemoryCompile(b, protocompile.SourceInfoExtraComments)
}

func BenchmarkGoogleapisMemoryCompileNoSourceInfo(b *testing.B) {
	benchmarkMemoryCompile(b, protocompile.SourceInfoNone)
}

func benchmarkMemoryCompile(b *testing.B, mode protocompile.SourceInfoMode) {
	sources := memorySources(b)
	resolver := protocompile.WithStandardImports(&protocompile.SourceResolver{
		Accessor: protocompile.SourceAccessorFromMap(sources),
	})
	b.ResetTimer()

	for range b.N {
		benchmarkProtocompile(b, &protocompile.Compiler{
			Resolver:       resolver,
			SourceInfoMode: mode,
		}, googleapisSources)
	}
}

func BenchmarkGoogleapisMemoryFastScan(b *testing.B) {
	sources := memorySources(b)
	var size int64
	for _, source := range sources {
		size += int64(len(source))
	}

	b.SetBytes(size)
	b.ResetTimer()

	for range b.N {
		for _, name := range googleapisSources {
			if _, err := fastscan.Scan(name, strings.NewReader(sources[name])); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func memorySources(b *testing.B) map[string]string {
	b.Helper()
	workspace, opener := googleapis.Get()
	sources := make(map[string]string, len(workspace.Paths()))
	for _, name := range workspace.Paths() {
		source, err := opener.Open(name)
		if err != nil {
			b.Fatal(err)
		}

		sources[name] = source.Text()
	}

	return sources
}
