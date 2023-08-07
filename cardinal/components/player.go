package components

import (
	"fmt"
	"strconv"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type PlayerComponent struct {
	PersonaTag string                       // username/player Persona tag
	Health     int                          // current player health (cap enforced in update loop)
	Coins      int                          // how much money the player has
	Weapon     storage.EntityID             // current player weapon; default is 0 for Melee
	Loc        types.Pair[float64, float64] // current location
	Dir        types.Pair[float64, float64] // array of movement directions with range [[-1,1],[-1,1]] where each types.Pair is the movement at a given timestep (divided uniformly over the tick) and the first direction is the one that determines player movement
	LastMove   types.Pair[float64, float64] // last player move; this must be a types.Pair of ints in [[-1,1],[-1,1]]
	IsRight    bool                         // whether player is facing right
	MoveNum    int                          // most recently-processed move
	Level      int                          // current player level
}

var Player = ecs.NewComponentType[PlayerComponent]()

func (p PlayerComponent) Simplify() types.BarePlayer {
	return types.BarePlayer{p.PersonaTag, p.Health, p.Coins, p.Loc.First, p.Loc.Second, p.IsRight, p.MoveNum, p.Level} // update Simplify for weapons & extraction point
}

func (p PlayerComponent) Testify(world *ecs.World) (types.TestPlayer, error) {
	weapon, err := Weapon.Get(world, p.Weapon)

	if err != nil {
		return types.TestPlayer{}, fmt.Errorf("Cardinal: error fetching player weapon", err)
	}

	return types.TestPlayer{p.PersonaTag, p.Health, p.Coins, weapon.Val, p.Loc.First, p.Loc.Second}, nil
}

func (p PlayerComponent) String(world *ecs.World) (string, error) {
	s := ""
	s += "PersonaTag: " + p.PersonaTag + "\n"
	s += "Health: " + strconv.Itoa(p.Health) + "\n"
	s += "Coins: " + strconv.Itoa(p.Coins) + "\n"
	weapon, err := Weapon.Get(world, p.Weapon)

	if err != nil {
		return "", fmt.Errorf("Cardinal: error fetching player weapon", err)
	}
	s += "Weapon: " + strconv.Itoa(int(weapon.Val)) + "\n"
	s += "Loc: " + strconv.FormatFloat(float64(p.Loc.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Loc.Second), 'e', -1, 32) + "\n"
	s += "Dir: " + strconv.FormatFloat(float64(p.Dir.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Dir.Second), 'e', -1, 32)

	return s, nil
}
