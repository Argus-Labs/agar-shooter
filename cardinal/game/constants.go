// Variables, weapons, and global constants
package game

import (
	"math"
	"sync"
	"time"

	"github.com/downflux/go-kd/kd"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type IWorldConstants struct {
	Weapons         	map[types.Weapon] types.WeaponData
	TickRate			int // Ticks per second
	ClientTickRate  	int // used to determine tickrate relative to cardinal server
	PlayerRadius    	float64 // used to determine which coins to collect
	PlayerSpeed     	int
	HealthPackValue  	int // amount of health each healthpack restores
	CoinRadius      	float64 // <= GameParams.CSize/2
	HealthRadius		float64 // radius of health packs
	HealthDensity		float64 // radius of health packs
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
			Melee: types.WeaponData{Attack: 10, Range: 4.0, MaxAmmo: -1, Reload: 200*time.Millisecond.Nanoseconds()},
			Slug: types.WeaponData{Attack: 3, Range: 6.9, MaxAmmo: 6, Reload: 5*time.Second.Nanoseconds()},
			TestWeapon: types.WeaponData{Attack: 2, Range: 4.0, MaxAmmo: -1, Reload: -1},
		},
		TickRate: 100,// ticks in ms
		ClientTickRate: 60,// used to determine tickrate relative to client
		PlayerRadius: 0.5,// used to determine which coins are close enough to collect
		PlayerSpeed: 2,
		HealthPackValue: 200,
		CoinRadius: 0.5,
		HealthRadius: 0.5,
		MaxCoinsPerTick: 100,
		MaxHealthPerTick: 100,
		HealthDensity: 0.025,// number of health packs per square unit
		MaxEntities: 4607704,
		InitRepeatSpawn: 10,
		BalanceFactor: 3,// multiple of min tree depth after which we should rebalance
	}

	GameParams				types.Game
	CoinMap					[][]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void // maps cells to sets of coin lists
	HealthMap				[][]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void // maps cells to sets of healthpack lists

	PlayerTree				= kd.New[*types.P](kd.O[*types.P]{ []*types.P{}, 2, 16, })
	PlayerCoins				= make(map[string] int)// the current number of coins each player has
	Players					= make(map[string] storage.EntityID) //players are personatags and components identified by strings; input into a map to make it easier to add and remove components
	LevelCoinParameters		= []float64{10, 10, 100}
	LevelHealthParameters	= []float64{950, 50, 4000}
	LevelAttackParameters	= []float64{-0.1, 0.1, 5}
	LevelCoins				= func(level int) int {
		return int(math.Min(LevelCoinParameters[0] + LevelCoinParameters[1]*float64(level), LevelCoinParameters[2]))
	}
	LevelHealth				= func(level int) int {
		return int(math.Min(LevelHealthParameters[0] + LevelHealthParameters[1]*float64(level), LevelHealthParameters[2]))
	}
	LevelAttack				= func(level int) float64 {
		return math.Min(LevelAttackParameters[0] + LevelAttackParameters[1]*float64(level), LevelAttackParameters[2])
	}
	Width, Height				int
	CoinMutex					= &sync.RWMutex{}
	HealthMutex					= &sync.RWMutex{}
	ClientView					= types.Pair[float64, float64]{First: 30, Second: 20} // client viewing window
	DefaultWeapon types.Weapon	= Melee
	Attacks						= make([]types.AttackTriple, 0)
	MaxCoinsInCell				= func() int {
		return int((GameParams.CSize * GameParams.CSize) / (3 * WorldConstants.CoinRadius * WorldConstants.CoinRadius * math.Pi))
	}
	MaxCoins					= func() int {
		return int(math.Min((float64(MaxCoinsInCell())*GameParams.Dims.First*GameParams.Dims.Second)/float64(GameParams.CSize*GameParams.CSize*2)+float64(3*len(Players)), float64(WorldConstants.MaxEntities-len(Players))))
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
