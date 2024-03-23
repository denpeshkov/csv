// Based on: https://github.com/golang/go/blob/master/src/encoding/csv/reader_test.go

package csv

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type readTest struct {
	Name      string
	Input     string
	Output    [][]string
	Positions [][][2]int
	Err       error

	// These fields are copied into the Reader
	Delimiter rune
	Comment   rune
}

var readTests = []readTest{{
	Name:   "Simple",
	Input:  "a,b,c\n",
	Output: [][]string{{"a", "b", "c"}},
}, {
	Name:   "Simple2",
	Input:  "a,b,c\nd,e",
	Output: [][]string{{"a", "b", "c"}, {"d", "e"}},
}, {
	Name:   "CRLF",
	Input:  "a,b\r\nc,d\r\n",
	Output: [][]string{{"a", "b"}, {"c", "d"}},
}, {
	Name:   "BareCR",
	Input:  "a,b\rc,d\r\n",
	Output: [][]string{{"a", "b"}, {"c", "d"}},
}, {
	Name: "RFC4180test",
	Input: `#field1,field2,field3
"aaa","bb
b","ccc"
"a,a","b""bb","ccc"
zzz,yyy,xxx
`,
	Output: [][]string{
		{"#field1", "field2", "field3"},
		{"aaa", "bb\nb", "ccc"},
		{"a,a", `b"bb`, "ccc"},
		{"zzz", "yyy", "xxx"},
	},
}, {
	Name:   "NoEOLTest",
	Input:  "a,b,c",
	Output: [][]string{{"a", "b", "c"}},
}, {
	Name:      "Semicolon",
	Input:     "a;b;c\n",
	Output:    [][]string{{"a", "b", "c"}},
	Delimiter: ';',
}, {
	Name: "MultiLine",
	Input: `"two
line","one line","three
line
field"`,
	Output: [][]string{{"two\nline", "one line", "three\nline\nfield"}},
}, {
	Name:  "BlankLine",
	Input: "a,b,c\n\nd,e,f\n\n",
	Output: [][]string{
		{"a", "b", "c"},
		{"d", "e", "f"},
	},
}, {
	Name:   "LeadingSpace",
	Input:  " a,  b,   c\n",
	Output: [][]string{{" a", "  b", "   c"}},
}, {
	Name:    "Comment",
	Input:   "#1,2,3\na,b,c\n#comment",
	Output:  [][]string{{"a", "b", "c"}},
	Comment: '#',
}, {
	Name:   "NoComment",
	Input:  "#1,2,3\na,b,c",
	Output: [][]string{{"#1", "2", "3"}, {"a", "b", "c"}},
}, {
	Name:   "TrailingDelimiterEOF",
	Input:  "a,b,c,",
	Output: [][]string{{"a", "b", "c", ""}},
}, {
	Name:   "TrailingDelimiterEOL",
	Input:  "a,b,c,\n",
	Output: [][]string{{"a", "b", "c", ""}},
}, {
	Name:   "TrailingDelimiterSpaceEOF",
	Input:  "a,b,c, ",
	Output: [][]string{{"a", "b", "c", " "}},
}, {
	Name:   "TrailingDelimiterSpaceEOL",
	Input:  "a,b,c, \n",
	Output: [][]string{{"a", "b", "c", " "}},
}, {
	Name:   "TrailingDelimiterLine3",
	Input:  "a,b,c\nd,e,f\ng,hi,",
	Output: [][]string{{"a", "b", "c"}, {"d", "e", "f"}, {"g", "hi", ""}},
}, {
	Name:   "NotTrailingDelimiter3",
	Input:  "a,b,c, \n",
	Output: [][]string{{"a", "b", "c", " "}},
}, {
	Name: "DelimiterFieldTest",
	Input: `x,y,z,w
x,y,z,
x,y,,
x,,,
,,,
"x","y","z","w"
"x","y","z",""
"x","y","",""
"x","","",""
"","","",""
`,
	Output: [][]string{
		{"x", "y", "z", "w"},
		{"x", "y", "z", ""},
		{"x", "y", "", ""},
		{"x", "", "", ""},
		{"", "", "", ""},
		{"x", "y", "z", "w"},
		{"x", "y", "z", ""},
		{"x", "y", "", ""},
		{"x", "", "", ""},
		{"", "", "", ""},
	},
}, {
	Name:  "TrailingDelimiterIneffective1",
	Input: "a,b,\nc,d,e",
	Output: [][]string{
		{"a", "b", ""},
		{"c", "d", "e"},
	},
}, {
	Name:  "CRLFInQuotedField", // Issue 21201
	Input: "A,\"Hello\r\nHi\",B\r\n",
	Output: [][]string{
		{"A", "Hello\r\nHi", "B"},
	},
}, {
	Name:   "TrailingCR",
	Input:  "field1,field2\r",
	Output: [][]string{{"field1", "field2"}},
}, {
	Name:   "QuotedTrailingCR",
	Input:  "\"field\"\r",
	Output: [][]string{{"field"}},
}, {
	Name:   "FieldCR",
	Input:  "field\rfield\r",
	Output: [][]string{{"field"}, {"field"}},
}, {
	Name:   "FieldCRCR",
	Input:  "field\r\rfield\r\r",
	Output: [][]string{{"field"}, {"field"}},
}, {
	Name:   "FieldCRCRLF",
	Input:  "field\r\r\nfield\r\r\n",
	Output: [][]string{{"field"}, {"field"}},
}, {
	Name:   "FieldCRCRLFCR",
	Input:  "field\r\r\n\rfield\r\r\n\r",
	Output: [][]string{{"field"}, {"field"}},
}, {
	Name:   "FieldCRCRLFCRCR",
	Input:  "field\r\r\n\r\rfield\r\r\n\r\r",
	Output: [][]string{{"field"}, {"field"}},
}, {
	Name:  "MultiFieldCRCRLFCRCR",
	Input: "field1,field2\r\r\n\r\rfield1,field2\r\r\n\r\r,",
	Output: [][]string{
		{"field1", "field2"},
		{"field1", "field2"},
		{"", ""},
	},
}, {
	Name:      "NonASCIIDelimiterAndComment",
	Input:     "a£b,c£ \td,e\n€ comment\n",
	Output:    [][]string{{"a", "b,c", " \td,e"}},
	Delimiter: '£',
	Comment:   '€',
}, {
	Name:      "NonASCIIDelimiterAndCommentWithQuotes",
	Input:     "a€\"  b,\"€ c\nλ comment\n",
	Output:    [][]string{{"a", "  b,", " c"}},
	Delimiter: '€',
	Comment:   'λ',
}, {
	// λ and θ start with the same byte.
	// This tests that the parser doesn't confuse such characters.
	Name:      "NonASCIIDelimiterConfusion",
	Input:     "\"abθcd\"λefθgh",
	Output:    [][]string{{"abθcd", "efθgh"}},
	Delimiter: 'λ',
	Comment:   '€',
}, {
	Name:    "NonASCIICommentConfusion",
	Input:   "λ\nλ\nθ\nλ\n",
	Output:  [][]string{{"λ"}, {"λ"}, {"λ"}},
	Comment: 'θ',
}, {
	Name:   "QuotedFieldMultipleLF",
	Input:  "\"\n\n\n\n\"",
	Output: [][]string{{"\n\n\n\n"}},
}, {
	Name:  "MultipleCRLF",
	Input: "\r\n\r\n\r\n\r\n",
}, {
	Name:    "HugeLines",
	Input:   strings.Repeat("#ignore\n", 10000) + "" + strings.Repeat("@", 5000) + "," + strings.Repeat("*", 5000),
	Output:  [][]string{{strings.Repeat("@", 5000), strings.Repeat("*", 5000)}},
	Comment: '#',
}, {
	Name:   "DoubleQuoteWithTrailingCRLF",
	Input:  "\"foo\"\"bar\"\r\n",
	Output: [][]string{{`foo"bar`}},
}, {
	Name:   "EvenQuotes",
	Input:  `""""""""`,
	Output: [][]string{{`"""`}},
},
	// Errors
	{
		Name:  "BadDoubleQuotes",
		Input: `a""b,c`,
		Err:   ErrBareQuote,
	}, {
		Name:  "BadBareQuote",
		Input: `a "word","b"`,
		Err:   ErrBareQuote,
	}, {
		Name:  "BadTrailingQuote",
		Input: `"a word",b"`,
		Err:   ErrBareQuote,
	}, {
		Name:  "ExtraneousQuote",
		Input: `"a "word","b"`,
		Err:   ErrQuote,
	}, {
		Name:  "StartLine1", // Issue 19019
		Input: "a,\"b\nc\"d,e",
		Err:   ErrQuote,
	}, {
		Name:   "StartLine2",
		Input:  "a,b\n\"d\n\n,e",
		Err:    ErrQuote,
		Output: [][]string{{"a", "b"}},
	}, {
		Name:   "QuotedTrailingCRCR",
		Input:  "\"field\"\r\r",
		Output: [][]string{{"field"}},
	}, {
		Name:  "QuoteWithTrailingCRLF",
		Input: "\"foo\"bar\"\r\n",
		Err:   ErrQuote,
	}, {
		Name:  "OddQuotes",
		Input: `"""""""`,
		Err:   ErrQuote,
	}}

func TestRead(t *testing.T) {
	newReader := func(tt readTest) (*Reader, string) {
		r := NewReader(strings.NewReader(tt.Input))

		if tt.Delimiter != 0 {
			r.Delimiter = tt.Delimiter
		}
		r.Comment = tt.Comment
		return r, tt.Input
	}

	for _, tt := range readTests {
		t.Run(tt.Name, func(t *testing.T) {
			r, _ := newReader(tt)
			out, err := r.ReadAll()

			if wantErr := tt.Err; wantErr != nil {
				if err != wantErr {
					t.Errorf("ReadAll() error mismatch:\ngot  %v (%#v)\nwant %v (%#v)", err, err, wantErr, wantErr)
				}
				if !cmp.Equal(out, tt.Output) {
					t.Errorf("ReadAll() output:\ngot  %q\nwant %q", out, tt.Output)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected Read() error: %v", err)
				}
				if !cmp.Equal(out, tt.Output) {
					t.Errorf("ReadAll() output:\ngot  %q\nwant %q", out, tt.Output)
				}
			}
		})
	}
}
