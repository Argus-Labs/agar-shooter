// Variables, weapons, and global constants
package game

import (
	"math"
	"sync"
	"time"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"

	"github.com/downflux/go-kd/kd"
)

type IWorldConstants struct {
	Weapons         	map[types.Weapon] types.WeaponData
	TickRate			int // Ticks per second
	ClientTickRate  	int // used to determine tickrate relative to cardinal server
	PlayerRadius    	float64 // used to determine which coins to collect
	HealthRadius		float64 // radius of health packs
	HealthDensity		float64 // radius of health packs
	PlayerSpeed     	int
	CoinRadius      	float64 // <= GameParams.CSize/2
	MaxCoinsPerTick		int
	MaxHealthPerTick	int
	MaxEntities			int
	InitRepeatSpawn 	int
	BalanceFactor		int// multiple of min tree depth after which we should rebalance; the higher, the fewer spikes in processing time at the cost of higher average processing time for lots of players
}

type IFooConstants struct {
	Foo string
}

const ( // add more weapons as needed
	Dud = iota - 1 // empty weapon
	Melee
	Slug
	TestWeapon
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
		Weapons: map[types.Weapon]types.WeaponData{
			Dud: types.WeaponData{Attack: 0, Range: 0.0, MaxAmmo: 0, Reload: 0},
			Melee: types.WeaponData{Attack: 2, Range: 4.0, MaxAmmo: -1, Reload: 1*time.Second.Nanoseconds()},
			Slug: types.WeaponData{Attack: 3, Range: 6.9, MaxAmmo: 6, Reload: 5*time.Second.Nanoseconds()},
			TestWeapon: types.WeaponData{Attack: 2, Range: 4.0, MaxAmmo: -1, Reload: -1},
		},
		TickRate: 10,// ticks per second
		ClientTickRate: 60,// used to determine tickrate relative to client
		PlayerRadius: 0.5,// used to determine which coins are close enough to collect
		PlayerSpeed: 2,
		CoinRadius: 0.5,
		HealthRadius: 0.5,
		MaxCoinsPerTick: 1000,
		HealthDensity: 0.1,// number of health packs per square unit
		MaxEntities: 4607704,
		InitRepeatSpawn: 1,
		BalanceFactor: 3,// multiple of min tree depth after which we should rebalance
	}

	GameParams		types.Game
	CoinMap    		= make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void) // maps cells to sets of coin lists
	HealthMap  		= make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void) // maps cells to sets of healthpack lists

	PlayerTree		= kd.New[*types.P](kd.O[*types.P]{ []*types.P{}, 2, 16, })
	PlayerCoins		= make(map[string] int)// the current number of coins each player has
	Players			= make(map[string] storage.EntityID) //players are names and components identified by strings; input into a map to make it easier to add and remove components
	LevelCoins		= map[int] int {// maps from level to number of coins to next level
		0: 0,
		1: 20,
		2: 30,
		3: 40,
		4: 50,
		5: 60,
		6: 70,
		7: 80,
		8: 90,
		9: 100,
	}
	LevelHealth		= map[int] int{// maps from level to max health
		0: 100,
		1: 110,
		2: 120,
		3: 130,
		4: 140,
		5: 150,
		6: 160,
		7: 170,
		8: 180,
		9: 190,
	}
	LevelAttack		= map[int] float64 {// maps from level to damage bonus
		0: 0,
		1: 0.05,
		2: 0.1,
		3: 0.15,
		4: 0.2,
		5: 0.25,
		6: 0.3,
		7: 0.35,
		8: 0.4,
		9: 0.45,
	}
	Width, Height				int
	CoinMutex					= &sync.RWMutex{}
	HealthMutex					= &sync.RWMutex{}
	ClientView					= types.Pair[float64, float64]{First: 30, Second: 20} // client viewing window
	DefaultWeapon types.Weapon	= Melee
	Attacks						= make([]types.AttackTriple, 0)
	MaxCoinsInCell				= func() int {
		return int(GameParams.CSize * GameParams.CSize / (3 * WorldConstants.CoinRadius * WorldConstants.CoinRadius * math.Pi))
	}
	MaxCoins					= func() int {
		return int(math.Min(float64(MaxCoinsInCell())*GameParams.Dims.First*GameParams.Dims.Second/GameParams.CSize/GameParams.CSize/4+float64(3*len(Players)), float64(WorldConstants.MaxEntities-len(Players))))
	}
	MaxHealth					= func() int {
		return int(math.Ceil(float64(GameParams.Dims.First * GameParams.Dims.Second) * WorldConstants.HealthDensity))
	}
	MaxHealthInCell				= func() int {
		return int(math.Max(1, math.Ceil(float64(MaxHealth()) / float64(GameParams.CSize * GameParams.CSize))))
	}
	TotalCoins					= 0
	TotalHealth					= 0
)
