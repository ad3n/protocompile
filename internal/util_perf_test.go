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

package internal

import (
	"strings"
	"testing"

	"github.com/ad3n/protocompile/internal/cases"
)

var utilResult string

func legacyJSONName(s string) string {
	return cases.Converter{Case: cases.Camel, NaiveSplit: true, NoLowercase: true}.Convert(s)
}

func legacyEscapeBytes[B ~string | ~[]byte](data B) string {
	buf := new(strings.Builder)
	for i := range len(data) {
		c := data[i]
		switch c {
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		case '"':
			buf.WriteString("\\\"")
		case '\'':
			buf.WriteString("\\'")
		case '\\':
			buf.WriteString("\\\\")
		default:
			if c >= 0x20 && c < 0x7f {
				buf.WriteByte(c)
				continue
			}

			buf.WriteRune('\\')
			buf.WriteByte('0' + ((c >> 6) & 0x7))
			buf.WriteByte('0' + ((c >> 3) & 0x7))
			buf.WriteByte('0' + (c & 0x7))
		}
	}

	return buf.String()
}

func FuzzUtilityCompatibility(f *testing.F) {
	var allBytes [256]byte
	for i := range allBytes {
		allBytes[i] = byte(i)
	}

	f.Add(string(allBytes[:]))
	for _, s := range []string{"", "plainField", "foo__bar_", "_élève_世界", "a_ȿ", "\xff_\xfe", "\n\r\t\\\"'", strings.Repeat("a", 256)} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		if got, want := JSONName(s), legacyJSONName(s); got != want {
			t.Fatalf("JSONName(%q) = %q, want %q", s, got, want)
		}

		want := legacyEscapeBytes(s)
		if got := EscapeBytes(s); got != want {
			t.Fatalf("EscapeBytes(%q) = %q, want %q", s, got, want)
		}

		data := []byte(s)
		got := EscapeBytes(data)
		for i := range data {
			data[i] = 0
		}

		if got != want {
			t.Fatalf("EscapeBytes([]byte(%q)) = %q, want %q after mutation", s, got, want)
		}
	})
}

func BenchmarkUtility(b *testing.B) {
	for _, tc := range []struct{ name, value string }{
		{"plain", "plainFieldName"},
		{"snake", "some_long_field_name_with_underscores"},
		{"unicode", "élève_世界_ȿ"},
		{"invalid", "x\xff_y\xfe"},
		{"escaped", strings.Repeat("hello\n\x00", 16)},
	} {
		b.Run("JSONName/"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				utilResult = JSONName(tc.value)
			}
		})
		b.Run("EscapeBytes/"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				utilResult = EscapeBytes(tc.value)
			}
		})
	}
}
