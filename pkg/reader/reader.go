package reader

import (
	"errors"
	"io"
)

const (
	// Note: Error types
	EndOfBuffer        = "EndOfBuffer"
	UnexpectedEOF      = "UnexpectedEOF"
	VarIntSizeExceeded = "VarIntSizeExceeded"
)

type Reader struct {
	buf  []byte
	next int
}

func FromBuffer(buf []byte) Reader {
	return Reader{
		buf:  buf,
		next: 0,
	}
}

// NOTE: Y-CRDT update v1 encoding format:
// 1. `clientsLen` | max 4 bytes
// 2. For each client:
//   - `blocksLen` | max 4 bytes
//   - `client` | max 4 bytes
//   - `clock` | max 4 bytes | starting clock for blocks
func DecodeUpdate(buf []byte) {
	r := FromBuffer(buf)
	clientsLen, _ := r.ReadU32()
	clients := make(map[uint32]string, clientsLen)

	for _ = range clientsLen {
		blocksLen, _ := r.ReadU32()
		client, _ := r.ReadU32()
		clock, _ := r.ReadU32()

		blocks := make([]string, blocksLen)
		for i := range blocksLen {

		}
	}
}

func (r *Reader) ReadU32() (uint32, error) {
	var num uint32 = 0
	var len uint = 0

	for {
		i, err := r.ReadU8()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		num |= uint32(i&0b01111111) << len
		len += 7

		if i < 0b10000000 {
			return num, nil
		}

		// a proper setting for 32bit int would be 35 bits, however for Yjs compatibility
		// we allow wrap up up to 64bit ints (with int overflow wrap)
		if len > 70 {
			return 0, errors.New(VarIntSizeExceeded)
		}
	}

	return 0, errors.New(UnexpectedEOF)
}

func (r *Reader) ReadU8() (uint8, error) {
	if r.next == len(r.buf) {
		return 0, errors.New(EndOfBuffer)
	}
	n := r.buf[r.next]
	r.next = r.next + 1
	return n, nil
}
