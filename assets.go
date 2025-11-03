package pregdk

import (
	_ "embed"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

var (
	//go:embed assets/required_item_list.json
	requiredItemList []byte
	//go:embed assets/vanilla_items.nbt
	itemRuntimeIDData []byte
	//go:embed assets/block_states.nbt
	blockStateData []byte
)

const (
	itemVersion        = 241
	blockVersion int32 = (1 << 24) | (21 << 16) | (60 << 8)
)

// itemData ...
func itemData() []byte {
	var items map[string]struct {
		RuntimeID      int32          `nbt:"runtime_id"`
		ComponentBased bool           `nbt:"component_based"`
		Version        int32          `nbt:"version"`
		Data           map[string]any `nbt:"data,omitempty"`
	}
	err := nbt.Unmarshal(itemRuntimeIDData, &items)
	if err != nil {
		panic(err)
	}
	var legacyItems = make(map[string]int32)
	for name, e := range items {
		legacyItems[name] = e.RuntimeID
	}
	data, err := nbt.Marshal(legacyItems)
	if err != nil {
		panic(err)
	}
	return data
}
