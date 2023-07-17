// Variables, weapons, and global constants
package game

import (
	"math"
	"sync"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

const ( // add more weapons as needed
	Dud = iota - 1 // empty weapon
	Melee
	Slug
)

var (
	GameParams types.Game
	World      = inmem.NewECSWorld()
	CoinMap    = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.void) // maps cells to sets of coin lists
	HealthMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.void)        // maps cells to sets of healthpack lists
	WeaponMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.void)        // maps cells to sets of weapon lists
	PlayerMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.void)        // maps cells to sets of player name-location types.Pairs

	Players       = make(map[string]storage.EntityID) //players are names and components identified by strings; input into a map to make it easier to add and remove components
	Width, Height int
	Weapons       = map[types.Weapon]types.WeaponData{
		Dud:   types.WeaponData{Attack: 0, Range: 0.0},
		Melee: types.WeaponData{Attack: 4, Range: 4.0},
		Slug:  types.WeaponData{Attack: 3, Range: 6.9},
	}
	mutex                       = &sync.RWMutex{}
	ClientView                  = types.Pair[float64, float64]{First: 30, Second: 20} // client viewing window
	DefaultWeapon  types.Weapon = Melee
	Attacks                     = make([]types.AttackTriple, 0)
	maxCoinsInCell              = func() int { return int(GameParams.CSize * GameParams.CSize / (3 * coinRadius * coinRadius * math.Pi)) }
	maxCoins                    = func() int {
		return int(math.Min(float64(maxCoinsInCell())*GameParams.Dims.First*GameParams.Dims.Second/GameParams.CSize/GameParams.CSize/4+float64(3*len(Players)), float64(MAXENTITIES-len(Players))))
	}
	totalCoins = 0
)

const (
	TickRate         = 5   // ticks per second
	ClientTickRate   = 60  // used to determine tickrate relative to cardinal server
	PlayerRadius     = 0.5 // used to determine which coins to collect
	ExtractionRadius = 10  // determines when players are in range of their extraction point
	sped             = 2   // player speed
	coinRadius       = 0.5 // <= GameParams.CSize/2
	maxCoinsPerTick  = 1000
	MAXENTITIES      = 4607704
	InitRepeatSpawn  = 1
)
