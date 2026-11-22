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

package slicesx

import (
	"iter"
)

func CutAfterFunc[S ~[]E, E any](s S, p func(int, E) bool) (before, after S, found bool) {
	idx := IndexFunc(s, p)
	if idx == -1 {
		return s, s[len(s):], false
	}
	return s[:idx+1], s[idx+1:], true
}

func SplitAfterFunc[S ~[]E, E any](s S, sep func(int, E) bool) iter.Seq[S] {
	return func(yield func(S) bool) {
		for {
			before, after, found := CutAfterFunc(s, sep)
			if !yield(before) {
				return
			}
			if !found {
				break
			}
			s = after
		}
	}
}
