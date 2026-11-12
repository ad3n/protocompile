package parser

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ad3n/protocompile/reporter"
)

func TestPooledParserPreservesAST(t *testing.T) {
	const source = `syntax = "proto3"; message First { string name = 1; }`
	handler := reporter.NewHandler(nil)
	lexer, err := newLexer(strings.NewReader(source), "first.proto", handler)
	if err != nil {
		t.Fatal(err)
	}

	protoParse(lexer)
	want, err := ResultFromAST(lexer.res, true, handler)
	if err != nil {
		t.Fatal(err)
	}

	first, err := Parse("first.proto", strings.NewReader(source), reporter.NewHandler(nil))
	if err != nil {
		t.Fatal(err)
	}

	for i, input := range []string{
		`syntax = "proto3"; message Broken {`,
		strings.Repeat("message Nested {", 100) + strings.Repeat("}", 100),
		`syntax = "proto3"; message Second { int64 value = 1; }`,
		"",
	} {
		_, err := Parse("other.proto", strings.NewReader(input), reporter.NewHandler(nil))
		if (err != nil) != (i == 0) {
			t.Fatalf("parse %d: unexpected error %v", i, err)
		}
	}

	got, err := ResultFromAST(first, true, reporter.NewHandler(nil))
	if err != nil {
		t.Fatal(err)
	}

	if !proto.Equal(want.FileDescriptorProto(), got.FileDescriptorProto()) {
		t.Fatal("parser reuse changed the retained AST")
	}
}
