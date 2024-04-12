package writer

type Writer struct {
	Buf []byte
}

func NewWriter() Writer {
	return Writer{
		Buf: []byte{},
	}
}

func (w *Writer) WriteU32(value uint32) {
	for value >= 0b10000000 {
		b := (value & 0b01111111) | 0b10000000
		w.WriteU8(uint8(b))
		value = value >> 7
	}
	w.WriteU8(uint8(value & 0b01111111))
}

func (w *Writer) WriteU8(value uint8) {
	w.Buf = append(w.Buf, value)
}
