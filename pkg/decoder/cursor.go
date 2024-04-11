package decoder

type Cursor struct {
	buf  []byte
	next int
}

func (c Cursor) hasContent() bool {
	return c.next != len(c.buf)
}
