package components

import (
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type WeaponComponent struct {
	Loc types.Pair[float64, float64]
	Val types.Weapon// weapon type
	Ammo int// number of attacks left
	LastAttack int64// time of last attack
}

var Weapon = ecs.NewComponentType[WeaponComponent]()
