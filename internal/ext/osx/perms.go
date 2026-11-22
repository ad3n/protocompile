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

// Package osx contains extensions to the os package.
package osx

// Go does not consider it worthy to include constants for the user permission
// bits, so we do so here.
//
// The constants are named for the names chmod gives them. U, G, and O mean
// "user", "group", and "other"; R, W, and X mean "read", "write", "execute".
const (
	// User read and write.
	PermUR = 0o400
	PermUW = 0o200

	// Group read and write.
	PermGR = 0o40
	PermGW = 0o20

	// World read and write.
	PermOR = 0o4
	PermOW = 0o2

	// All read and write.
	PermAR = PermUR | PermGR | PermOR
	PermAW = PermUW | PermGW | PermOW
)
