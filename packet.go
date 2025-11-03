package pregdk

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// pk ...
type pk struct {
	id      uint32
	marshal func(io protocol.IO)
}

func (pk pk) ID() uint32             { return pk.id }
func (pk pk) Marshal(io protocol.IO) { pk.marshal(io) }

// packetF ...
func packetF(id uint32, marshal func(io protocol.IO)) []packet.Packet {
	return []packet.Packet{pk{id: id, marshal: marshal}}
}

// packetFunc ...
func packetFunc(pk packet.Packet) func() packet.Packet {
	return func() packet.Packet {
		return pk
	}
}
