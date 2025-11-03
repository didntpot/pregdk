package pregdk

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// animatePacket ...
type animatePacket struct {
	ActionType      int32
	EntityRuntimeID uint64
	RowingTime      float32
}

func (pk *animatePacket) ID() uint32 { return packet.IDAnimate }
func (pk *animatePacket) Marshal(io protocol.IO) {
	io.Varint32(&pk.ActionType)
	io.Varuint64(&pk.EntityRuntimeID)
	if pk.ActionType == packet.AnimateActionRowLeft || pk.ActionType == packet.AnimateActionRowRight {
		io.Float32(&pk.RowingTime)
	}
}
