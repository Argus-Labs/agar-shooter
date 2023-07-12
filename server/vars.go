// Variables, weapons, and global constants
package main

import (
	"sync"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
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
	HealthMap				= make(map[Pair[int, int]] map[Pair[storage.EntityID, Pair[float64,float64]]] void)// maps cells to sets of healthpack lists
	WeaponMap				= make(map[Pair[int, int]] map[Pair[storage.EntityID, Pair[float64,float64]]] void)// maps cells to sets of weapon lists
	PlayerMap				= make(map[Pair[int,int]] map[Pair[storage.EntityID, Pair[float64,float64]]] void)// maps cells to sets of player name-location pairs
	PlayerComp				= ecs.NewComponentType[PlayerComponent]()
	CoinComp				= ecs.NewComponentType[CoinComponent]()
	HealthComp				= ecs.NewComponentType[HealthComponent]()
	WeaponComp				= ecs.NewComponentType[HealthComponent]()
	Players					= make(map[string] storage.EntityID)//players are names and components identified by strings; input into a map to make it easier to add and remove components
	MoveTx					= ecs.NewTransactionType[Move]()//(World, "move")
	Width, Height			int
	Weapons					= map[Weapon] WeaponData{
								Dud: WeaponData{0, 0.0},
								Melee: WeaponData{4, 4.0},
								Slug: WeaponData{3, 6.9},
							}
	globalMut				= &sync.RWMutex{}
	ClientView				= Pair[float64,float64]{30,20}// client viewing window
	DefaultWeapon Weapon	= Melee
	Attacks					= make([]AttackTriple, 0)
)

const (
	TickRate			= 5// ticks per second
	ClientTickRate		= 60// used to determine tickrate relative to cardinal server
	PlayerRadius		= 0.5// used to determine which coins to collect
	ExtractionRadius	= 10// determines when players are in range of their extraction point
	sped				= 2// player speed
)
