package parser

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/ad3n/protocompile/ast"
	"github.com/ad3n/protocompile/reporter"
)

func TestPooledParserConcurrentRetainedAST(t *testing.T) {
	const workers = 16
	const iterations = 64

	files := make([]*ast.FileNode, workers*iterations)
	errors := make(chan error, workers)
	var group sync.WaitGroup
	for worker := range workers {
		group.Go(func() {
			for iteration := range iterations {
				index := worker*iterations + iteration
				name := fmt.Sprintf("M%d", index)
				source := `syntax = "proto3"; message ` + name + ` { string value = 1; }`
				file, err := Parse("test.proto", strings.NewReader(source), reporter.NewHandler(nil))
				if err != nil {
					errors <- err
					return
				}

				files[index] = file
				_, err = Parse("bad.proto", strings.NewReader("message Broken {"), reporter.NewHandler(nil))
				if err == nil {
					errors <- fmt.Errorf("malformed input unexpectedly accepted")
					return
				}
			}
		})
	}

	group.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}

	if t.Failed() {
		return
	}

	for i, file := range files {
		result, err := ResultFromAST(file, true, reporter.NewHandler(nil))
		if err != nil {
			t.Fatal(err)
		}

		if got := result.FileDescriptorProto().GetMessageType()[0].GetName(); got != fmt.Sprintf("M%d", i) {
			t.Fatalf("retained AST %d changed to %q", i, got)
		}
	}
}

func TestPooledParserAfterReporterPanic(t *testing.T) {
	for range 32 {
		func() {
			defer func() {
				if got := recover(); got != "reporter panic" {
					t.Fatalf("unexpected panic: %v", got)
				}
			}()

			handler := reporter.NewHandler(reporter.NewReporter(func(reporter.ErrorWithPos) error {
				panic("reporter panic")
			}, nil))
			_, _ = Parse("bad.proto", strings.NewReader("message Broken {"), handler)
		}()

		file, err := Parse("good.proto", strings.NewReader(`syntax = "proto3"; message Good {}`), reporter.NewHandler(nil))
		if err != nil {
			t.Fatal(err)
		}

		result, err := ResultFromAST(file, true, reporter.NewHandler(nil))
		if err != nil {
			t.Fatal(err)
		}

		if result.FileDescriptorProto().GetMessageType()[0].GetName() != "Good" {
			t.Fatal("parser state leaked across panic")
		}
	}
}
