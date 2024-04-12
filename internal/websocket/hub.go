package websocket

import (
	"github.com/rejdeboer/multiplayer-server/internal/sync"
)

type Hub struct {
	Rooms map[*Room]bool
}

func NewHub() *Hub {
	return &Hub{
		Rooms: make(map[*Room]bool),
	}
}

func (h *Hub) GetDocumentRoom(doc *sync.Doc) *Room {
	var docRoom *Room
	for room := range h.Rooms {
		if room.Doc.ID == doc.ID {
			docRoom = room
		}
	}

	if docRoom == nil {
		docRoom = NewRoom(doc)
		go docRoom.Run(h)
		h.Rooms[docRoom] = true
	}

	return docRoom
}
