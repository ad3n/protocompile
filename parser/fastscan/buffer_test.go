package fastscan

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestScanBufferBoundaries(t *testing.T) {
	for _, padding := range []int{0, 1, 4093, 4094, 4095, 4096, 8191} {
		input := "/*" + strings.Repeat("x", padding) + "世界\t**/\n//" + strings.Repeat("y", padding) + "\xff\t世界\npackage example.long_name; import \"世界.proto\";"
		for _, reader := range []io.Reader{
			strings.NewReader(input),
			iotest.OneByteReader(strings.NewReader(input)),
			iotest.DataErrReader(strings.NewReader(input)),
		} {
			result, err := Scan("boundary.proto", reader)
			if err != nil {
				t.Fatalf("padding %d: %v", padding, err)
			}

			if result.PackageName != "example.long_name" || len(result.Imports) != 1 || result.Imports[0].Path != "世界.proto" {
				t.Fatalf("padding %d: unexpected result %+v", padding, result)
			}
		}
	}
}

func TestLexerBufferPositions(t *testing.T) {
	input := "//" + strings.Repeat("x", 4095) + "世界\t\n/*" + strings.Repeat("a", 4095) + "世界\t*/\n\tpackage"
	lexer := newLexer(strings.NewReader(input))
	token, text, err := lexer.Lex()
	if err != nil || token != identifierToken || text != "package" {
		t.Fatalf("unexpected token %v %q %v", token, text, err)
	}

	if lexer.prevTokenLine != 2 || lexer.prevTokenCol != 8 || lexer.curLine != 2 || lexer.curCol != 15 {
		t.Fatalf("unexpected position: previous %d:%d, current %d:%d", lexer.prevTokenLine, lexer.prevTokenCol, lexer.curLine, lexer.curCol)
	}

	lexer = newLexer(strings.NewReader("//" + strings.Repeat("x", 4095) + "世界\t"))
	token, _, err = lexer.Lex()
	if err != nil || token != eofToken || lexer.curLine != 0 || lexer.curCol != 4104 {
		t.Fatalf("unexpected comment EOF: %v %v at %d:%d", token, err, lexer.curLine, lexer.curCol)
	}
}

func FuzzBufferedRuneReader(f *testing.F) {
	for _, input := range []string{"", "ascii\t\n", "世界", "\xff\xc0\x80\xe4\xb8", strings.Repeat("a", 4095) + "世界"} {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		for mode := range 3 {
			reader := func() io.Reader {
				r := strings.NewReader(input)
				switch mode {
				case 1:
					return iotest.OneByteReader(r)
				case 2:
					return iotest.TimeoutReader(r)
				}

				return r
			}
			got := &runeReader{rr: bufio.NewReader(reader())}
			want := bufio.NewReader(reader())
			for i := 0; ; i++ {
				gr, ge := got.readRune()
				wr, _, we := want.ReadRune()
				if gr != wr || (ge == nil) != (we == nil) {
					t.Fatalf("rune %d: got %U %v, want %U %v", i, gr, ge, wr, we)
				}

				if ge != nil {
					if ge.Error() != we.Error() {
						t.Fatalf("got error %v, want %v", ge, we)
					}

					break
				}

				if i%7 == 0 {
					got.unreadRune(gr)
					again, err := got.readRune()
					if again != wr || err != nil {
						t.Fatal("unread changed the rune")
					}
				}
			}
		}
	})
}
