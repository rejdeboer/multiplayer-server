package sync

import (
	"github.com/google/uuid"
)

type Doc struct {
	ID          uuid.UUID
	StateVector []byte
}
