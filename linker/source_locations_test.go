package linker

import (
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
)

func TestSourceLocationChains(t *testing.T) {
	protos := []*descriptorpb.SourceCodeInfo_Location{
		{Path: []int32{4, 0}, Span: []int32{0, 1, 2}},
		{Path: []int32{4, 1}, Span: []int32{1, 2, 3, 4}},
		{Path: []int32{4, 0}, Span: []int32{4, 5, 6}},
		{Path: []int32{4, 0}, Span: []int32{7, 8, 9}},
		{Path: nil, Span: []int32{10, 11, 12}},
		{Path: []int32{}, Span: []int32{13, 14, 15}},
	}
	locations, index := asSourceLocations(protos)
	for i, next := range []int{2, 0, 3, 0, 5, 0} {
		if locations[i].Next != next {
			t.Fatalf("location %d: next %d, want %d", i, locations[i].Next, next)
		}
	}

	for _, first := range []int{0, 1, 4} {
		if got := index[pathKey(protos[first].Path)]; got != first {
			t.Fatalf("first location: got %d, want %d", got, first)
		}
	}

	if locations[1].EndLine != 3 || locations[1].EndColumn != 4 {
		t.Fatal("multiline span changed")
	}

	empty, emptyIndex := asSourceLocations(nil)
	if len(empty) != 0 || len(emptyIndex) != 0 {
		t.Fatal("empty source locations changed")
	}
}
