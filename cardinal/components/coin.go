package components

import "github.com/argus-labs/new-game/types"

type CoinComponent struct {
	Loc types.Pair[float64, float64]
	Val int // how many coins the component represents; could represent different denominations as larger or different colored coins
}
