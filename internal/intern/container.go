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

package intern

import "github.com/ad3n/protocompile/internal/ext/mapsx"

// Set is a set of intern IDs.
type Set map[ID]struct{}

// ContainsID returns whether s contains the given ID.
func (s Set) ContainsID(id ID) bool {
	_, ok := s[id]
	return ok
}

// AddID adds an ID to s, and returns whether it was added.
func (s Set) AddID(id ID) (inserted bool) {
	return mapsx.AddZero(s, id)
}

// Add adds a string to s, and returns whether it was added.
func (s Set) Add(table *Table, key string) (inserted bool) {
	k := table.Intern(key)
	_, ok := s[k]
	if !ok {
		s[k] = struct{}{}
	}
	return !ok
}

// Map is a map keyed by intern IDs.
type Map[T any] map[ID]T

// AddID adds an ID to m, and returns whether it was added.
func (m Map[T]) AddID(id ID, v T) (mapped T, inserted bool) {
	return mapsx.Add(m, id, v)
}
