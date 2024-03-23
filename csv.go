package csv

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

// The errors that can be returned during parsing.
var (
	ErrBareQuote = errors.New("bare \" in non-quoted-field")
	ErrQuote     = errors.New("extraneous or missing \" in quoted-field")
)

// The errors that can be returned during Reader setup.
var (
	ErrInvalidQuote   = errors.New("invalid quotation")
	ErrInvalidDelim   = errors.New("invalid delimiter")
	ErrInvalidComment = errors.New("invalid comment")
)

func valid(r rune) bool {
	return r != '\r' && r != '\n' && utf8.ValidRune(r) && r != utf8.RuneError
}

// A Reader reads records from a CSV document.
// CSV document must be valid UTF-8 encoded.
// A Reader treats (\r), (\n), (\r\n) as a single record terminator.
// Quote, Delimiter and Comment must all be different runes.
type Reader struct {
	// Quote is the field quotation. By default (").
	// Quote must be a valid rune and must not be (\r), (\n).
	Quote rune
	// Delimiter is the field delimiter. By default (,).
	// Delimiter must be a valid rune and must not be (\r), (\n).
	Delimiter rune
	// Comment is the comment character if not 0. By default 0.
	// Comment must be a valid rune and must not be (\r), (\n).
	Comment rune

	r      *bufio.Reader
	field  bytes.Buffer
	err    error
	record []string
}

// NewReader returns a [Reader] that reads a CSV document from r.
//   - Quote is set to (")";
//   - Delimiter is set to (,);
//   - Comment is set to 0 (no comments).
func NewReader(r io.Reader) *Reader {
	return &Reader{
		Quote:     '"',
		Delimiter: ',',
		Comment:   0,
		r:         bufio.NewReader(r),
	}
}

// SetQuote sets the quotation.
func (r *Reader) SetQuote(quote rune) error {
	if r.Quote == r.Delimiter || r.Quote == r.Comment || !valid(r.Quote) || r.Quote == 0 {
		return ErrInvalidQuote
	}
	r.Quote = quote
	return nil
}

// SetDelimiter sets the delimiter.
func (r *Reader) SetDelimiter(delimiter rune) error {
	if r.Delimiter == r.Quote || r.Delimiter == r.Comment || !valid(r.Delimiter) || r.Delimiter == 0 {
		return ErrInvalidDelim
	}
	r.Delimiter = delimiter
	return nil
}

// SetComment sets the comment.
func (r *Reader) SetComment(comment rune) error {
	if r.Comment == r.Quote || r.Comment == r.Delimiter || !valid(r.Comment) {
		return ErrInvalidComment
	}
	r.Comment = comment
	return nil
}

// Read reads one record from r.
func (r *Reader) Read() (record []string, err error) {
	for s := startLine(r); s != nil; {
		s = s(r)
	}
	record = r.record
	err = r.err

	r.field.Reset()
	r.err = nil
	r.record = nil
	return record, err
}

// ReadAll reads all the remaining records from r.
func (r *Reader) ReadAll() (records [][]string, err error) {
	var record []string
	for {
		record, err = r.Read()
		if err != nil {
			break
		}
		if record != nil {
			records = append(records, record)
		}
		if _, err := r.r.Peek(1); err == io.EOF {
			break
		}
	}
	return records, err
}

// next returns the next rune in the input.
func (r *Reader) next() (rune, error) {
	ch, _, err := r.r.ReadRune()
	if err == io.EOF {
		return eof, nil
	}
	if err != nil {
		return 0, err
	}
	return ch, nil
}

// backup steps back one rune in the input.
func (r *Reader) backup() error {
	return r.r.UnreadRune()
}

// write appends one rune to the current field.
func (r *Reader) write(rune rune) error {
	if _, err := r.field.WriteRune(rune); err != nil {
		return err
	}
	return nil
}

// endField takes the current field, adds it to the current record and clears the current field so parsing can continue.
func (r *Reader) endField() {
	r.record = append(r.record, r.field.String())
	r.field.Reset()
}
