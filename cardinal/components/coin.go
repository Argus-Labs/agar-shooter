package components

import (
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type CoinComponent struct {
	Loc types.Pair[float64, float64]
	Val int // how many coins the component represents; could represent different denominations as larger or different colored coins
}

var Coin = ecs.NewComponentType[CoinComponent]()
