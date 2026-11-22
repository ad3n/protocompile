// Copyright 2020-2026 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fdp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ad3n/protocompile/experimental/source"
)

func TestSourceLocationSlices(t *testing.T) {
	t.Parallel()

	file := source.NewFile("test.proto", "abc\ndef")
	prefix := []int32{4, 0}
	suffix := []int32{2, 1}
	location := newSourceLocation(source.Span{File: file, End: 3}, prefix, suffix)

	prefix[0], suffix[0] = 9, 9
	require.Equal(t, []int32{4, 0, 2, 1}, location.Path)
	require.Equal(t, []int32{0, 0, 3}, location.Span)

	require.Equal(t, []int32{0, 0, 3, 9}, append(location.Span, 9))
	require.Equal(t, []int32{4, 0, 2, 1}, location.Path)

	root := newSourceLocation(source.Span{File: file, End: 7}, nil, nil)
	require.Equal(t, []int32{0, 0, 1, 3}, root.Span)
	require.Nil(t, root.Path)
}
