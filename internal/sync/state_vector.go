package sync

import "github.com/rejdeboer/multiplayer-server/pkg/reader"

type StateVector = map[uint32]uint32

func DecodeStateVector(buf []byte) StateVector {
	r := reader.FromBuffer(buf)

	svLen, _ := r.ReadU32()
	sv := make(StateVector, svLen)

	for _ = range svLen {
		client, _ := r.ReadU32()
		clock, _ := r.ReadU32()
		sv[client] = clock
	}

	return sv
}
