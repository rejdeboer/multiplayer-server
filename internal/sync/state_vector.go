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

func diffStateVectors(localSv StateVector, remoteSv StateVector) map[uint32]uint32 {
	diff := make(map[uint32]uint32)
	for client, remoteClock := range remoteSv {
		localClock := localSv[client]
		if localClock > remoteClock {
			diff[client] = remoteClock
		}
	}
	for client, _ := range localSv {
		if remoteSv[client] == 0 {
			diff[client] = 0
		}
	}
	return diff
}
