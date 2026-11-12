package linker

import (
	"fmt"
	"testing"
)

type dependencyTestFile struct {
	File
	path string
}

func (f *dependencyTestFile) Path() string {
	return f.path
}

func TestDependencyIndex(t *testing.T) {
	first := &dependencyTestFile{path: "first.proto"}
	duplicate := &dependencyTestFile{path: "first.proto"}
	second := &dependencyTestFile{path: "second.proto"}
	files := Files{first, second, duplicate}
	index := newDependencyIndex(files)
	for _, path := range []string{"first.proto", "second.proto", "missing.proto", ""} {
		if got, want := index.FindFileByPath(path), files.FindFileByPath(path); got != want {
			t.Fatalf("lookup %q: got %v, want %v", path, got, want)
		}
	}

	if got := newDependencyIndex(nil).FindFileByPath("missing.proto"); got != nil {
		t.Fatalf("empty lookup: got %v", got)
	}
}

func BenchmarkDependencyLookup(b *testing.B) {
	for _, size := range []int{1, 4, 8, 16, 64} {
		files := make(Files, size)
		paths := make([]string, size+1)
		for i := range files {
			paths[i] = fmt.Sprintf("google/example/dependency_%d.proto", i)
			files[i] = &dependencyTestFile{path: paths[i]}
		}

		paths[size] = "google/example/missing.proto"
		index := newDependencyIndex(files)
		b.Run(fmt.Sprintf("%d/linear", size), func(b *testing.B) {
			for i := range b.N {
				path := paths[i%len(paths)]
				got := files.FindFileByPath(path)
				if (got == nil) != (path == paths[size]) {
					b.Fatal("unexpected lookup result")
				}
			}
		})
		b.Run(fmt.Sprintf("%d/indexed", size), func(b *testing.B) {
			for i := range b.N {
				path := paths[i%len(paths)]
				got := index.FindFileByPath(path)
				if (got == nil) != (path == paths[size]) {
					b.Fatal("unexpected lookup result")
				}
			}
		})
	}
}
