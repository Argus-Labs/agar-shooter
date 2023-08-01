// Variables, weapons, and global constants
package game

import (
	"math"
	"sync"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type IWorldConstants struct {
	Dud              types.WeaponData // empty weapon, iota - 1
	Melee            types.WeaponData // weapon two
	Slug             types.WeaponData // weapon three
	Weapons          map[types.Weapon]types.WeaponData
	TickRate         int     // Ticks per second
	ClientTickRate   int     // used to determine tickrate relative to cardinal server
	PlayerRadius     float64 // used to determine which coins to collect
	ExtractionRadius int     // determines when players are in range of their extraction point
	PlayerSpeed      int
	CoinRadius       float64 // <= GameParams.CSize/2
	MaxCoinsPerTick  int
	MaxEntities      int
	InitRepeatSpawn  int
}

type IFooConstants struct {
	Foo string
}

const ( // add more weapons as needed
	Dud = iota - 1 // empty weapon
	Melee
	Slug
)

var (
	ExposedConstants = []types.IConstant{
		{
			Label: "world",
			Value: WorldConstants,
		},
	}

	// WorldConstants is a public constant that can be queried through `query_constant`
	// because it is in the list of ExposedConstants
	WorldConstants = IWorldConstants{
		Dud:   types.WeaponData{Attack: 0, Range: 0.0},
		Melee: types.WeaponData{Attack: 4, Range: 4.0},
		Slug:  types.WeaponData{Attack: 3, Range: 6.9},
		Weapons: map[types.Weapon]types.WeaponData{
			Dud:   types.WeaponData{Attack: 0, Range: 0.0},
			Melee: types.WeaponData{Attack: 4, Range: 4.0},
			Slug:  types.WeaponData{Attack: 3, Range: 6.9},
		},
		TickRate:         10,
		ClientTickRate:   60,
		PlayerRadius:     0.5,
		ExtractionRadius: 10,
		PlayerSpeed:      2,
		CoinRadius:       0.5,
		MaxCoinsPerTick:  1000,
		MaxEntities:      4607704,
		InitRepeatSpawn:  1,
	}

	// FooConstant is a private constant that cannot be queried through `query_constant`
	// because it is not in the list of ExposedConstants
	//FooConstants = IFooConstants{
	//	Foo: "Bar",
	//}

	GameParams types.Game
	CoinMap    = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void) // maps cells to sets of coin lists
	HealthMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.Void)        // maps cells to sets of healthpack lists
	WeaponMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.Void)        // maps cells to sets of weapon lists
	PlayerMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.Void)        // maps cells to sets of player name-location types.Pairs

	Players        = make(map[string]storage.EntityID) //players are names and components identified by strings; input into a map to make it easier to add and remove components
	Width, Height  int
	Mutex          = &sync.RWMutex{}
	ClientView     = types.Pair[float64, float64]{First: 30, Second: 20} // client viewing window
	Attacks        = make([]types.AttackTriple, 0)
	MaxCoinsInCell = func() int {
		return int(GameParams.CSize * GameParams.CSize / (3 * WorldConstants.CoinRadius * WorldConstants.CoinRadius * math.Pi))
	}
	MaxCoins = func() int {
		return int(math.Min(float64(MaxCoinsInCell())*GameParams.Dims.First*GameParams.Dims.Second/GameParams.CSize/GameParams.CSize/4+float64(3*len(Players)), float64(WorldConstants.MaxEntities-len(Players))))
	}
	TotalCoins = 0
)
