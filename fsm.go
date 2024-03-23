package csv

import "io"

type stateFn func(r *Reader) stateFn

const eof rune = -1

func startLine(r *Reader) stateFn {
	c, err := r.next()
	if err != nil {
		r.err = err
		return nil
	}

	switch c {
	case eof:
		r.err = io.EOF
		fallthrough
	case '\r', '\n':
		return nil
	case r.Comment:
		return comment
	default:
		if err := r.backup(); err != nil {
			r.err = err
			return nil
		}
		return startField
	}
}

func comment(r *Reader) stateFn {
	for {
		c, err := r.next()
		if err != nil {
			r.err = err
			return nil
		}

		switch c {
		case '\r', '\n':
			return startLine
		case eof:
			return nil
		}
	}
}

func startField(r *Reader) stateFn {
	c, err := r.next()
	if err != nil {
		r.err = err
		return nil
	}

	switch c {
	case eof:
		r.err = io.EOF
		fallthrough
	case '\r', '\n':
		r.endField()
		return nil
	case r.Quote:
		return quotedField
	default:
		if err := r.backup(); err != nil {
			r.err = err
			return nil
		}
		return field
	}
}

func field(r *Reader) stateFn {
	for {
		c, err := r.next()
		if err != nil {
			r.err = err
			return nil
		}

		switch c {
		case eof:
			r.err = io.EOF
			fallthrough
		case '\r', '\n':
			r.endField()
			return nil
		case r.Delimiter:
			r.endField()
			return startField
		case r.Quote:
			r.err = ErrBareQuote
			return nil
		default:
			r.appendRune(c)
		}
	}
}

func quotedField(r *Reader) stateFn {
	for {
		c, err := r.next()
		if err != nil {
			r.err = err
			return nil
		}

		switch c {
		case eof:
			r.err = ErrQuote
			return nil
		case r.Quote:
			return doubleQuotedField
		default:
			r.appendRune(c)
		}
	}
}

func doubleQuotedField(r *Reader) stateFn {
	c, err := r.next()
	if err != nil {
		r.err = err
		return nil
	}

	switch c {
	case eof:
		r.err = io.EOF
		fallthrough
	case '\r', '\n':
		r.endField()
		return nil
	case r.Quote:
		r.appendRune(c)
		return quotedField
	case r.Delimiter:
		r.endField()
		return startField
	default:
		r.err = ErrQuote
		return nil
	}
}
