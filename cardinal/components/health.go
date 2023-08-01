package components

import (
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type HealthComponent struct {
	Loc types.Pair[float64, float64] // location
	Val int                          // how much health it contains
}

var Health = ecs.NewComponentType[HealthComponent]()
