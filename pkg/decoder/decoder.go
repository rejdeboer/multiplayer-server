package decoder

import (
	"errors"
	"io"
)

// NOTE: Y-CRDT v1 encoding format:
// 1. `clients_len` - max 4 bytes

type Decoder struct {
	buf  []byte
	next int
}

func From(buf []byte) Decoder {
	return Decoder{
		buf:  buf,
		next: 0,
	}
}

func (d *Decoder) readU32() (uint32, error) {
	var num uint32 = 0
	var len uint = 0

	for {
		r, err := d.readU8()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		num |= uint32(r&0b01111111) << len
		len += 7

		if r < 0b10000000 {
			return num, nil
		}

		// a proper setting for 32bit int would be 35 bits, however for Yjs compatibility
		// we allow wrap up up to 64bit ints (with int overflow wrap)
		if len > 70 {
			return 0, errors.New("VarIntSizeExceeded")
		}
	}

	return 0, errors.New("UnexpectedEOF")
}

func (d *Decoder) readU8() (uint8, error) {
	if d.next == len(d.buf) {
		return 0, errors.New("EndOfBuffer")
	}
	n := d.buf[d.next]
	d.next = d.next + 1
	return n, nil
}
