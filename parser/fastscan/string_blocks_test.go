package fastscan

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestStringLiteralBlocks(t *testing.T) {
	for _, size := range []int{0, 1, 15, 16, 4093, 4094, 4095, 4096, 8192} {
		prefix := strings.Repeat("a", size)
		for _, quote := range []string{"\"", "'"} {
			for _, suffix := range []string{"", "世界", "\xff", "\t\n", "\\n\\x41\\101", "\\x1q", "\\z", "\x00\r"} {
				input := quote + prefix + suffix + quote
				want := prefix
				switch suffix {
				case "\xff":
					want += "\ufffd"
				case "\\n\\x41\\101":
					want += "\nAA"
				case "\\x1q":
					want += "\x01q"
				default:
					want += suffix
				}

				for _, reader := range []io.Reader{strings.NewReader(input), iotest.OneByteReader(strings.NewReader(input)), iotest.DataErrReader(strings.NewReader(input))} {
					lexer := newLexer(reader)
					token, got, err := lexer.Lex()
					if err != nil || token != stringToken || got != want {
						t.Fatalf("size %d quote %q suffix %q: token %v, value %q, error %v", size, quote, suffix, token, got, err)
					}
				}
			}
		}
	}
}

func TestStringLiteralBlockEOFAndPosition(t *testing.T) {
	payload := strings.Repeat("a", 12000)
	lexer := newLexer(strings.NewReader("\"" + payload))
	token, got, err := lexer.Lex()
	if err != nil || token != stringToken || got != payload || lexer.curCol != len(payload)+1 {
		t.Fatal("unterminated ASCII string changed")
	}

	lexer = newLexer(strings.NewReader("\"" + payload + "\n\t世界\" next"))
	_, _, err = lexer.Lex()
	if err != nil || lexer.curLine != 1 || lexer.curCol != 11 {
		t.Fatalf("unexpected string position %d:%d, error %v", lexer.curLine, lexer.curCol, err)
	}

	token, got, err = lexer.Lex()
	if err != nil || token != identifierToken || got != "next" || lexer.prevTokenLine != 1 || lexer.prevTokenCol != 12 {
		t.Fatal("position after string changed")
	}
}

func FuzzStringLiteralBlocks(f *testing.F) {
	for _, input := range []string{"", "abc\"", "world\\n世界\"", "\\x1q\xff\t\n", strings.Repeat("a", 4095) + "世界\""} {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 32768 {
			t.Skip()
		}

		for _, quote := range []string{"\"", "'"} {
			whole := newLexer(strings.NewReader(quote + input))
			fragmented := newLexer(iotest.OneByteReader(strings.NewReader(quote + input)))
			wt, ws, we := whole.Lex()
			ft, fs, fe := fragmented.Lex()
			if wt != ft || ws != fs || (we == nil) != (fe == nil) || whole.curLine != fragmented.curLine || whole.curCol != fragmented.curCol {
				t.Fatal("fragmenting the reader changed a string literal or its position")
			}
		}
	})
}
