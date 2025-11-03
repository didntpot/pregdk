package pregdk

import (
	"github.com/didntpot/multiversion/multiversion/mapping"
	"github.com/didntpot/multiversion/multiversion/mapping/chunk"
	"github.com/didntpot/multiversion/multiversion/mapping/translator"
	"github.com/didntpot/multiversion/multiversion/protocols/latest"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// proto ...
type proto struct {
	itemMapping  mapping.Item
	blockMapping mapping.Block

	itemTranslator  translator.ItemTranslator
	blockTranslator translator.BlockTranslator

	current minecraft.Protocol
}

// Protocol ...
func Protocol(proxy bool) minecraft.Protocol {
	itemMapping := mapping.NewItemMapping(itemData(), requiredItemList, itemVersion, !proxy)
	blockMapping := mapping.NewBlockMapping(blockStateData)
	latestBlockMapping := latest.NewBlockMapping()
	return proto{
		itemTranslator: translator.NewItemTranslator(
			itemMapping,
			latest.NewItemMapping(!proxy),
			blockMapping, latestBlockMapping,
		),
		blockTranslator: translator.NewBlockTranslator(
			blockMapping, latestBlockMapping,
			chunk.NewNetworkPersistentEncoding(blockMapping, blockVersion),
			chunk.NewBlockPaletteEncoding(blockMapping, blockVersion),
			false, false,
		),

		current: minecraft.DefaultProtocol,
	}
}

func (proto proto) ID() int32   { return 844 }
func (proto proto) Ver() string { return "1.21.111" }

func (proto proto) Packets(listener bool) packet.Pool {
	pool := proto.current.Packets(listener)
	pool[packet.IDAnimate] = packetFunc(&animatePacket{})
	return pool
}
func (proto proto) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return proto.current.NewReader(r, shieldID, enableLimits)
}
func (proto proto) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return proto.current.NewWriter(w, shieldID)
}

func (proto proto) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pkt := pk.(type) {
	case *animatePacket:
		return single(&packet.Animate{
			ActionType:      pkt.ActionType,
			EntityRuntimeID: pkt.EntityRuntimeID,
			RowingTime:      pkt.RowingTime,
		})
	}
	return proto.itemTranslator.UpgradeItemPackets(single(pk), conn)
}
func (proto proto) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pkt := pk.(type) {
	case *packet.ActorEvent:
		if pkt.EventType > packet.ActorEventDrinkMilk {
			return nil
		}
	case *packet.Animate:
		return packetF(packet.IDAnimate, func(io protocol.IO) {
			io.Varint32(&pkt.ActionType)
			io.Varuint64(&pkt.EntityRuntimeID)
			if pkt.ActionType == packet.AnimateActionRowLeft || pkt.ActionType == packet.AnimateActionRowRight {
				io.Float32(&pkt.RowingTime)
			}
		})
	case *packet.BiomeDefinitionList:
		return nil
	case *packet.CameraInstruction:
		return packetF(packet.IDCameraInstruction, func(io protocol.IO) {
			protocol.OptionalMarshaler(io, &pkt.Set)
			protocol.OptionalFunc(io, &pkt.Clear, io.Bool)
			protocol.OptionalMarshaler(io, &pkt.Fade)
			protocol.OptionalMarshaler(io, &pkt.Target)
			protocol.OptionalFunc(io, &pkt.RemoveTarget, io.Bool)
			protocol.OptionalMarshaler(io, &pkt.FieldOfView)
		})
	case *packet.ShowStoreOffer:
		return packetF(packet.IDShowStoreOffer, func(io protocol.IO) {
			io.String(pointer(pkt.OfferID.String()))
			io.Uint8(&pkt.Type)
		})
	default:
		if pk.ID() > packet.IDServerBoundPackSettingChange {
			return nil
		}
	}
	return proto.itemTranslator.DowngradeItemPackets([]packet.Packet{pk}, conn)
}

// pointer ...
func pointer[Y any](x Y) *Y {
	return &x
}

// single ...
func single[T packet.Packet](pk T) []packet.Packet {
	return []packet.Packet{pk}
}
