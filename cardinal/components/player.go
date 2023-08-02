package components

import (
	"strconv"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type PlayerComponent struct {
	PersonaTag string                       // username; ip for now
	Health     int                          // current player health (cap enforced in update loop)
	Coins      int                          // how much money the player has
	Weapon     types.Weapon                 // current player weapon; default is 0 for Melee
	Loc        types.Pair[float64, float64] // current location
	Dir        types.Pair[float64, float64] // array of movement directions with range [[-1,1],[-1,1]] where each types.Pair is the movement at a given timestep (divided uniformly over the tick) and the first direction is the one that determines player movement
	LastMove   types.Pair[float64, float64] // last player move; this must be a types.Pair of ints in [[-1,1],[-1,1]]
	Extract    types.Pair[float64, float64] // extraction point; as long as the player is within some distance of the extraction point, player coins are offloaded
	IsRight    bool                         // whether player is facing right
	MoveNum    int                          // most recently-processed move
}

var Player = ecs.NewComponentType[PlayerComponent]()

func (p PlayerComponent) Simplify() types.BarePlayer {
	return types.BarePlayer{p.PersonaTag, p.Health, p.Coins, p.Loc.First, p.Loc.Second, p.IsRight, p.MoveNum} // update Simplify for weapons & extraction point
}

func (p PlayerComponent) Testify() types.TestPlayer {
	return types.TestPlayer{p.PersonaTag, p.Health, p.Coins, p.Weapon, p.Extract.First, p.Extract.Second, p.Loc.First, p.Loc.Second}
}

func (p PlayerComponent) String() string {
	s := ""
	s += "Name: " + p.PersonaTag + "\n"
	s += "Health: " + strconv.Itoa(p.Health) + "\n"
	s += "Coins: " + strconv.Itoa(p.Coins) + "\n"
	s += "Weapon: " + strconv.Itoa(int(p.Weapon)) + "\n"
	s += "Loc: " + strconv.FormatFloat(float64(p.Loc.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Loc.Second), 'e', -1, 32) + "\n"
	s += "Dir: " + strconv.FormatFloat(float64(p.Dir.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Dir.Second), 'e', -1, 32)

	return s
}
