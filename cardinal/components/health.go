package components

import "github.com/argus-labs/new-game/types"

type HealthComponent struct {
	Loc types.Pair[float64, float64] // location
	Val int                          // how much health it contains
}
