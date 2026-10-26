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

package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/bufbuild/protocompile/reporter"
)

const poolingMaxCapacity = 64 << 10

type (
	poolingVariant struct {
		name string
		read func(*protoLex, *bytes.Buffer) (string, error)
	}

	poolingWorkload struct {
		name   string
		inputs []string
	}
)

func poolingVariants() []poolingVariant {
	pool := &sync.Pool{New: func() any { return new(bytes.Buffer) }}
	return []poolingVariant{
		{"before", func(l *protoLex, _ *bytes.Buffer) (string, error) {
			var buf bytes.Buffer
			return poolingReadStringLiteral(l, '"', &buf)
		}},
		{"reuse", func(l *protoLex, _ *bytes.Buffer) (string, error) {
			return l.readStringLiteral('"')
		}},
		{"pool", func(l *protoLex, _ *bytes.Buffer) (string, error) {
			buf := pool.Get().(*bytes.Buffer)
			defer func() {
				buf.Reset()
				if buf.Cap() <= poolingMaxCapacity {
					pool.Put(buf)
				}
			}()

			return poolingReadStringLiteral(l, '"', buf)
		}},
	}
}

func poolingWorkloads() []poolingWorkload {
	return []poolingWorkload{
		{"empty", []string{`""`}},
		{"short", []string{`"hello world"`}},
		{"escaped", []string{`"hello\nworld\t\x41\101\u00e9\U0001f600"`}},
		{"unicode", []string{`"selamat pagi 世界 😊"`}},
		{"large", []string{`"` + strings.Repeat("abcdef", 1024) + `"`}},
		{"mixed", []string{`""`, `"name"`, `"path\nvalue"`, `"` + strings.Repeat("a", 1024) + `"`}},
		{"oversized", []string{`"` + strings.Repeat("a", poolingMaxCapacity+1) + `"`, `"small"`}},
	}
}

func BenchmarkLiteralBuffer(b *testing.B) {
	for _, workload := range poolingWorkloads() {
		b.Run(workload.name, func(b *testing.B) {
			inputs := make([][]byte, len(workload.inputs))
			for i, s := range workload.inputs {
				inputs[i] = []byte(s)
			}

			for _, variant := range poolingVariants() {
				b.Run(variant.name, func(b *testing.B) {
					b.Run("serial", func(b *testing.B) {
						var rr runeReader
						l := &protoLex{input: &rr}
						var scratch bytes.Buffer
						var result string
						b.ReportAllocs()
						b.ResetTimer()
						for i := 0; i < b.N; i++ {
							rr = runeReader{data: inputs[i%len(inputs)], pos: 1}
							var err error
							result, err = variant.read(l, &scratch)
							if err != nil {
								b.Fatal(err)
							}
						}

						b.StopTimer()
						runtime.KeepAlive(result)
					})
					b.Run("parallel", func(b *testing.B) {
						b.ReportAllocs()
						b.RunParallel(func(pb *testing.PB) {
							var rr runeReader
							l := &protoLex{input: &rr}
							var scratch bytes.Buffer
							var result string
							i := 0
							for pb.Next() {
								rr = runeReader{data: inputs[i%len(inputs)], pos: 1}
								var err error
								result, err = variant.read(l, &scratch)
								if err != nil {
									b.Error(err)
									return
								}

								i++
							}

							runtime.KeepAlive(result)
						})
					})
				})
			}
		})
	}
}

func TestLiteralBufferCompatibility(t *testing.T) {
	inputs := []string{`"\q\z"`, `"\xg"`, `"\777"`, `"\u12"`, `"\Uffffffff"`, "\"bad\nline\"", "\"\xff\xfe\"", "\"\x00\"", `"unterminated`, `"ends with\`}
	for _, workload := range poolingWorkloads() {
		inputs = append(inputs, workload.inputs...)
	}

	for _, variant := range poolingVariants() {
		t.Run(variant.name, func(t *testing.T) {
			t.Parallel()
			var scratch bytes.Buffer
			var results, expected []string
			for _, input := range inputs {
				run := func(read func(*protoLex, *bytes.Buffer) (string, error)) (string, string, int, []string) {
					var diagnostics []string
					handler := reporter.NewHandler(reporter.NewReporter(func(err reporter.ErrorWithPos) error {
						diagnostics = append(diagnostics, err.Error())
						return nil
					}, nil))
					l, err := newLexer(strings.NewReader(input), "pooling.proto", handler)
					if err != nil {
						t.Fatal(err)
					}

					l.input.pos = 1
					value, err := read(l, &scratch)
					return value, fmt.Sprint(err), l.input.pos, diagnostics
				}

				want, wantErr, wantPos, wantDiagnostics := run(func(l *protoLex, _ *bytes.Buffer) (string, error) {
					var buf bytes.Buffer
					return poolingReadStringLiteral(l, '"', &buf)
				})
				got, gotErr, gotPos, gotDiagnostics := run(variant.read)
				if got != want || gotErr != wantErr || gotPos != wantPos || !slices.Equal(gotDiagnostics, wantDiagnostics) {
					t.Fatalf("input %q: got (%q, %q, %d, %v), want (%q, %q, %d, %v)", input, got, gotErr, gotPos, gotDiagnostics, want, wantErr, wantPos, wantDiagnostics)
				}

				results = append(results, got)
				expected = append(expected, want)
			}

			scratch.WriteString(strings.Repeat("overwrite", 1024))
			if !slices.Equal(results, expected) {
				t.Fatal("returned strings changed after buffer reuse")
			}
		})
	}
}

func poolingReadStringLiteral(l *protoLex, quote rune, buf *bytes.Buffer) (string, error) {
	var escapeError reporter.ErrorWithPos
	var noMoreErrors bool
	reportErr := func(msg, badEscape string) {
		if noMoreErrors {
			return
		}

		if escapeError != nil {
			_, ok := l.addSourceError(escapeError)
			if !ok {
				noMoreErrors = true
			}
		}

		var err error
		if strings.HasSuffix(msg, "%s") {
			err = fmt.Errorf(msg, badEscape)
		} else {
			err = errors.New(msg)
		}

		escapeError = l.errWithCurrentPos(err, -len(badEscape))
	}

	for {
		c, _, err := l.input.readRune()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}

			return "", err
		}

		if c == '\n' {
			return "", errors.New("encountered end-of-line before end of string literal")
		}

		if c == quote {
			break
		}

		if c == 0 {
			reportErr("null character ('\\0') not allowed in string literal", string(rune(0)))
			continue
		}

		if c == '\\' {
			c, _, err = l.input.readRune()
			if err != nil {
				return "", err
			}

			switch {
			case c == 'x' || c == 'X':
				c1, sz1, err := l.input.readRune()
				if err != nil {
					return "", err
				}

				if c1 == quote || c1 == '\\' {
					l.input.unreadRune(sz1)
					reportErr("invalid hex escape: %s", "\\"+string(c))
					continue
				}

				c2, sz2, err := l.input.readRune()
				if err != nil {
					return "", err
				}

				var hex string
				if (c2 < '0' || c2 > '9') && (c2 < 'a' || c2 > 'f') && (c2 < 'A' || c2 > 'F') {
					l.input.unreadRune(sz2)
					hex = string(c1)
				} else {
					hex = string([]rune{c1, c2})
				}

				i, err := strconv.ParseInt(hex, 16, 32)
				if err != nil {
					reportErr("invalid hex escape: %s", "\\"+string(c)+hex)
					continue
				}

				buf.WriteByte(byte(i))
			case c >= '0' && c <= '7':
				c2, sz2, err := l.input.readRune()
				if err != nil {
					return "", err
				}

				var octal string
				if c2 < '0' || c2 > '7' {
					l.input.unreadRune(sz2)
					octal = string(c)
				} else {
					c3, sz3, err := l.input.readRune()
					if err != nil {
						return "", err
					}

					if c3 < '0' || c3 > '7' {
						l.input.unreadRune(sz3)
						octal = string([]rune{c, c2})
					} else {
						octal = string([]rune{c, c2, c3})
					}
				}

				i, err := strconv.ParseInt(octal, 8, 32)
				if err != nil {
					reportErr("invalid octal escape: %s", "\\"+octal)
					continue
				}

				if i > 0xff {
					reportErr("octal escape is out range, must be between 0 and 377: %s", "\\"+octal)
					continue
				}

				buf.WriteByte(byte(i))
			case c == 'u':
				u := make([]rune, 4)
				for i := range u {
					c2, sz2, err := l.input.readRune()
					if err != nil {
						return "", err
					}

					if c2 == quote || c2 == '\\' {
						l.input.unreadRune(sz2)
						u = u[:i]
						break
					}

					u[i] = c2
				}

				codepointStr := string(u)
				if len(u) < 4 {
					reportErr("invalid unicode escape: %s", "\\u"+codepointStr)
					continue
				}

				i, err := strconv.ParseInt(codepointStr, 16, 32)
				if err != nil {
					reportErr("invalid unicode escape: %s", "\\u"+codepointStr)
					continue
				}

				buf.WriteRune(rune(i))
			case c == 'U':
				u := make([]rune, 8)
				for i := range u {
					c2, sz2, err := l.input.readRune()
					if err != nil {
						return "", err
					}

					if c2 == quote || c2 == '\\' {
						l.input.unreadRune(sz2)
						u = u[:i]
						break
					}

					u[i] = c2
				}

				codepointStr := string(u)
				if len(u) < 8 {
					reportErr("invalid unicode escape: %s", "\\U"+codepointStr)
					continue
				}

				i, err := strconv.ParseInt(string(u), 16, 32)
				if err != nil {
					reportErr("invalid unicode escape: %s", "\\U"+codepointStr)
					continue
				}

				if i > 0x10ffff || i < 0 {
					reportErr("unicode escape is out of range, must be between 0 and 0x10ffff: %s", "\\U"+codepointStr)
					continue
				}

				buf.WriteRune(rune(i))
			case c == 'a':
				buf.WriteByte('\a')
			case c == 'b':
				buf.WriteByte('\b')
			case c == 'f':
				buf.WriteByte('\f')
			case c == 'n':
				buf.WriteByte('\n')
			case c == 'r':
				buf.WriteByte('\r')
			case c == 't':
				buf.WriteByte('\t')
			case c == 'v':
				buf.WriteByte('\v')
			case c == '\\':
				buf.WriteByte('\\')
			case c == '\'':
				buf.WriteByte('\'')
			case c == '"':
				buf.WriteByte('"')
			case c == '?':
				buf.WriteByte('?')
			default:
				reportErr("invalid escape sequence: %s", "\\"+string(c))
				continue
			}
		} else {
			buf.WriteRune(c)
		}
	}

	if escapeError != nil {
		return "", escapeError
	}

	return buf.String(), nil
}

func BenchmarkParseStringLiterals(b *testing.B) {
	for _, count := range []int{1, 32, 256} {
		var source strings.Builder
		source.WriteString(`syntax = "proto2"; message Example {`)
		for i := range count {
			fmt.Fprintf(&source, `optional string field_%d = %d [default = "hello\nworld\tvalue"];`, i, i+1)
		}

		source.WriteString("}")
		input := source.String()
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(input)))
			for b.Loop() {
				node, err := Parse("benchmark.proto", strings.NewReader(input), reporter.NewHandler(nil))
				if err != nil {
					b.Fatal(err)
				}

				runtime.KeepAlive(node)
			}
		})
	}
}

func TestStringLiteralBufferReuse(t *testing.T) {
	t.Parallel()
	inputs := []string{
		`"first value"`, `"second value"`, `""`, `"\q"`, `"\u12"`,
		`"` + strings.Repeat("a", maxStringBufferCapacity+1) + `"`,
		`"after oversized"`,
		`"` + strings.Repeat("a", maxStringBufferCapacity+1),
		`"after error"`,
	}
	l := &protoLex{}
	var results, expected []string
	for _, input := range inputs {
		fresh, err := newLexer(strings.NewReader(input), "reuse.proto", reporter.NewHandler(nil))
		if err != nil {
			t.Fatal(err)
		}

		l.input = fresh.input
		l.info = fresh.info
		l.handler = fresh.handler
		l.input.pos = 1
		var original bytes.Buffer
		want, wantErr := poolingReadStringLiteral(fresh, '"', &original)
		l.input.pos = 1
		l.input.err = nil
		got, gotErr := l.readStringLiteral('"')
		if got != want || fmt.Sprint(gotErr) != fmt.Sprint(wantErr) {
			t.Fatalf("result mismatch for input length %d: (%q, %v), want (%q, %v)", len(input), got, gotErr, want, wantErr)
		}

		if l.stringBuf.Len() != 0 || l.stringBuf.Cap() > maxStringBufferCapacity {
			t.Fatalf("buffer not released correctly: len=%d cap=%d", l.stringBuf.Len(), l.stringBuf.Cap())
		}

		if len(input) > maxStringBufferCapacity && l.stringBuf.Cap() != 0 {
			t.Fatal("oversized backing buffer retained")
		}

		results = append(results, got)
		expected = append(expected, want)
	}

	if !slices.Equal(results, expected) {
		t.Fatal("retained strings changed after reuse")
	}
}
