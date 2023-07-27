// Variables, weapons, and global constants
package main

import (
	"sync"
	"math"
	"time"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"

	"github.com/downflux/go-kd/kd"
)

const (// add more weapons as needed
	Dud = iota - 1// empty weapon
	Melee 
	Slug
)
var (
	GameParams				Game
	World					= inmem.NewECSWorld()
	CoinMap					= make(map[Pair[int, int]] map[Pair[storage.EntityID, Triple[float64,float64,int]]] void)// maps cells to sets of coin lists
	HealthMap				= make(map[Pair[int, int]] map[Pair[storage.EntityID, Triple[float64,float64,int]]] void)// maps cells to sets of healthpack lists
	WeaponMap				= make(map[Pair[int, int]] map[Pair[storage.EntityID, Triple[float64,float64,int]]] void)// maps cells to sets of weapon lists
	PlayerTree				= kd.New[*P](kd.O[*P]{ []*P{}, 2, 16, })
	PlayerComp				= ecs.NewComponentType[PlayerComponent]()
	CoinComp				= ecs.NewComponentType[CoinComponent]()
	HealthComp				= ecs.NewComponentType[HealthComponent]()
	WeaponComp				= ecs.NewComponentType[WeaponComponent]()
	PlayerMaxCoins			= make(map[string] int)// stores max coins achieved for each player
	Players					= make(map[string] storage.EntityID)//players are names and components identified by strings; input into a map to make it easier to add and remove components
	MoveTx					= ecs.NewTransactionType[Move]()//(World, "move")
	Width, Height			int
	Weapons					= map[Weapon] WeaponData{
								Dud: WeaponData{0, 0.0, 0, 0},
								Melee: WeaponData{2, 4.0, -1, 1*time.Second.Nanoseconds()},// last number is weapon cooldown in nanoseconds
								Slug: WeaponData{3, 6.9, 6, 5*time.Second.Nanoseconds()},
							}
	coinMutex				= &sync.RWMutex{}
	healthMutex				= &sync.RWMutex{}
	ClientView				= Pair[float64,float64]{30,20}// client viewing window
	DefaultWeapon Weapon	= Melee
	Attacks					= make([]AttackTriple, 0)
	maxCoinsInCell			= func() int { return int(GameParams.CSize*GameParams.CSize/(3*coinRadius*coinRadius*math.Pi)) }
	maxCoins				= func() int { return int(math.Min(float64(maxCoinsInCell())*GameParams.Dims.First*GameParams.Dims.Second/GameParams.CSize/GameParams.CSize/4 + float64(3*len(Players)), float64(MAXENTITIES - len(Players))))}
	totalCoins				= 0
	maxHealth				= func() int { return int(math.Ceil(float64(GameParams.Dims.First*GameParams.Dims.Second)*healthDensity)) }
	maxHealthInCell			= func() int { return int(math.Max(1, math.Ceil(float64(maxHealth())/float64(GameParams.CSize*GameParams.CSize)))) }
	totalHealth				= 0
	ExtractionCooldown		= 2*time.Second.Nanoseconds()// determines when players are in range of their extraction point
)

const (
	TickRate			= 10// ticks per second
	ClientTickRate		= 60// used to determine tickrate relative to cardinal server
	PlayerRadius		= 0.5// used to determine which coins to collect
	ExtractionRadius	= 10// determines when players are in range of their extraction point
	sped				= 2// player speed
	coinRadius			= 0.5// <= GameParams.CSize/2
	healthRadius		= 0.5// <= GameParams.CSize/2
	maxCoinsPerTick		= 1000
	healthDensity		= 0.1// number of health packs per square unit
	maxHealthPerTick	= 10
	MAXENTITIES			= 4607704
	InitRepeatSpawn		= 1
	balanceFactor		= 3// multiple of min tree depth after which we should rebalance; the higher, the fewer spikes in processing time there will be at the cost of higher average processing time for lots of players
)
