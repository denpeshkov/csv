package csv

type stateFn func(r *Reader) stateFn

const eof rune = -1

func startLine(r *Reader) stateFn {
	c, err := r.next()
	if err != nil {
		return stateErr(err)
	}
	switch c {
	case eof, '\r', '\n':
		return nil
	case r.Comment:
		return comment
	default:
		r.backup()
		return startField
	}
}

func comment(r *Reader) stateFn {
	for {
		c, err := r.next()
		if err != nil {
			return stateErr(err)
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
		return stateErr(err)
	}
	switch c {
	case eof, '\r', '\n':
		r.endField()
		return nil
	case r.Quote:
		return quotedField
	default:
		r.backup()
		return field
	}
}

func field(r *Reader) stateFn {
	for {
		c, err := r.next()
		if err != nil {
			return stateErr(err)
		}
		switch c {
		case eof, '\r', '\n':
			r.endField()
			return nil
		case r.Delimiter:
			r.endField()
			return startField
		case r.Quote:
			return stateErr(ErrBareQuote)
		}
		r.write(c)
	}
}

func quotedField(r *Reader) stateFn {
	for {
		c, err := r.next()
		if err != nil {
			return stateErr(err)
		}
		switch c {
		case eof:
			return stateErr(ErrQuote)
		case r.Quote:
			return doubleQuotedField
		}
		r.write(c)
	}
}

func doubleQuotedField(r *Reader) stateFn {
	c, err := r.next()
	if err != nil {
		return stateErr(err)
	}
	switch c {
	case eof, '\r', '\n':
		r.endField()
		return nil
	case r.Quote:
		r.write(c)
		return quotedField
	case r.Delimiter:
		r.endField()
		return startField
	}
	return stateErr(ErrQuote)
}

func stateErr(err error) stateFn {
	return func(r *Reader) stateFn {
		r.err = err
		return nil
	}
}
