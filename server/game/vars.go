// Variables, weapons, and global constants
package game

import (
	"github.com/argus-labs/world-engine/game/sample_game_server/server/component"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/types"
	"math"
	"sync"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"

	"github.com/downflux/go-kd/kd"
)

const ( // add more weapons as needed
	Dud = iota - 1 // empty weapon
	Melee
	Slug
)

var (
	GameParams     types.Game
	World          = inmem.NewECSWorld()
	CoinMap        = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void) // maps cells to sets of coin lists
	HealthMap      = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.Void)        // maps cells to sets of healthpack lists
	WeaponMap      = make(map[types.Pair[int, int]]map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.Void)        // maps cells to sets of weapon lists
	PlayerTree     = kd.New[*types.P](kd.O[*types.P]{[]*types.P{}, 2, 16})
	PlayerComp     = ecs.NewComponentType[component.PlayerComponent]()
	CoinComp       = ecs.NewComponentType[component.CoinComponent]()
	HealthComp     = ecs.NewComponentType[component.HealthComponent]()
	WeaponComp     = ecs.NewComponentType[component.HealthComponent]()
	PlayerMaxCoins = make(map[string]int)                 // used only for demo purposes TODO: remove after demo
	Players        = make(map[string]storage.EntityID)    //players are names and components identified by strings; input into a map to make it easier to add and remove components
	MoveTx         = ecs.NewTransactionType[types.Move]() //(World, "move")
	Width, Height  int
	Weapons        = map[types.Weapon]types.WeaponData{
		Dud:   types.WeaponData{0, 0.0},
		Melee: types.WeaponData{2, 4.0},
		Slug:  types.WeaponData{3, 6.9},
	}
	mutex                       = &sync.RWMutex{}
	ClientView                  = types.Pair[float64, float64]{30, 20} // client viewing window
	DefaultWeapon  types.Weapon = Melee
	Attacks                     = make([]types.AttackTriple, 0)
	maxCoinsInCell              = func() int { return int(GameParams.CSize * GameParams.CSize / (3 * coinRadius * coinRadius * math.Pi)) }
	maxCoins                    = func() int {
		return int(math.Min(float64(maxCoinsInCell())*GameParams.Dims.First*GameParams.Dims.Second/GameParams.CSize/GameParams.CSize/4+float64(3*len(Players)), float64(MAXENTITIES-len(Players))))
	}
	totalCoins = 0
)

const (
	TickRate         = 10  // ticks per second
	ClientTickRate   = 60  // used to determine tickrate relative to cardinal server
	PlayerRadius     = 0.5 // used to determine which coins to collect
	ExtractionRadius = -1  // determines when players are in range of their extraction point
	Sped             = 2   // player speed
	coinRadius       = 0.5 // <= GameParams.CSize/2
	maxCoinsPerTick  = 1000
	MAXENTITIES      = 4607704
	InitRepeatSpawn  = 1
	BalanceFactor    = 3 // multiple of min tree depth after which we should rebalance; the higher, the fewer spikes in processing time there will be at the cost of higher average processing time for lots of players
)