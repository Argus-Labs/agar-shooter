package components

import (
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type WeaponComponent struct {
	Loc types.Pair[float64, float64]
	Val types.Weapon // weapon type; TODO: implement ammo later outside of weapon component
	// cooldown, ammo, damage, range
}

var Weapon = ecs.NewComponentType[WeaponComponent]()