// Variables, weapons, and global constants
package game

import (
	"sync"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type IWorldConstants struct {
	Dud              types.WeaponData // empty weapon, iota - 1
	Melee            types.WeaponData // weapon two
	Slug             types.WeaponData // weapon three
	TickRate         int              // Ticks per second
	ClientTickRate   int              // used to determine tickrate relative to cardinal server
	PlayerRadius     float32          // used to determine which coins to collect
	ExtractionRadius int              // determines when players are in range of their extraction point
	PlayerSpeed      int
	CoinRadius       float32 // <= GameParams.CSize/2
	MaxCoinsPerTick  int
	MaxEntities      int
	InitRepeatSpawn  int
}

type IFooConstants struct {
	Foo string
}

var (
	// If you want the constant to be queryable through `query_constant`,
	// make sure to add the constant to the list of exposed constants
	ExposedConstants = []types.IConstant{
		{
			Label: "world",
			Value: WorldConstants,
		},
	}

	// WorldConstants is a public constant that can be queried through `query_constant`
	// because it is in the list of ExposedConstants
	WorldConstants = IWorldConstants{
		Dud:              types.WeaponData{Attack: 0, Range: 0.0},
		Melee:            types.WeaponData{Attack: 4, Range: 4.0},
		Slug:             types.WeaponData{Attack: 3, Range: 6.9},
		TickRate:         5,
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
	FooConstants = IFooConstants{
		Foo: "Bar",
	}

	GameParams types.Game
	CoinMap    = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.void) // maps cells to sets of coin lists
	HealthMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.void)        // maps cells to sets of healthpack lists
	WeaponMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.void)        // maps cells to sets of weapon lists
	PlayerMap  = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.void)        // maps cells to sets of player name-location types.Pairs

	Players       = make(map[string]storage.EntityID) //players are names and components identified by strings; input into a map to make it easier to add and remove components
	Width, Height int
	// Weapons       = map[types.Weapon]types.WeaponData{
	// 	Dud:   types.WeaponData{Attack: 0, Range: 0.0},
	// 	Melee: types.WeaponData{Attack: 4, Range: 4.0},
	// 	Slug:  types.WeaponData{Attack: 3, Range: 6.9},
	// }
	mutex                      = &sync.RWMutex{}
	ClientView                 = types.Pair[float64, float64]{First: 30, Second: 20} // client viewing window
	DefaultWeapon types.Weapon = Melee
	Attacks                    = make([]types.AttackTriple, 0)

	totalCoins = 0
)
