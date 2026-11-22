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

func (r *Queue[E]) Front() *E {
	if r.start == r.end {
		return nil
	}

	return &r.buf[r.start]
}

func (r *Queue[E]) Back() *E {
	if r.start == r.end {
		return nil
	}

	return &r.buf[(r.end-1)&(len(r.buf)-1)]
}

func (r *Queue[E]) PopBack() (E, bool) {
	if r.start == r.end {
		var z E
		return z, false
	}

	r.end--
	r.end &= len(r.buf) - 1
	return Take(r.buf, r.end)
}

func (r *Queue[_]) Clear() {
	clear(r.buf)
	r.start, r.end = 0, 0
	r.buf = r.buf[:0]
}
