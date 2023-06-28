package main

import (
	"strconv"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type void struct{}
var pewp void

type Pair[T1 any, T2 any] struct {
	First T1
	Second T2
}

type ItemComponent struct {}

type HealthComponent struct {
	Loc Pair[float64, float64]// location
	Val int// how much health it contains
}

type CoinComponent struct {
	Loc Pair[float64, float64]
	Val int// how many coins the component represents; could represent different denominations as larger or different colored coins
}

type Weapon int

const (// add more weapons as needed
	Melee = iota
	Sniper
)

type WeaponComponent struct {
	Loc Pair[float64, float64]
	Val Weapon// weapon type; TODO: implement ammo later outside of weapon component
	// cooldown, ammo, damage, range
}

type ItemMapComponent struct {// TODO: maybe don't turn these into components
	Items map[Pair[int, int]] map[Pair[storage.EntityID, Pair[float64,float64]]] void// maps cells to sets of item lists
}

type PlayerMapComponent struct {
	Players map[Pair[int,int]] map[Pair[storage.EntityID, Pair[float64,float64]]] void// maps cells to sets of player name-location pairs
}

type Direction struct {
	Theta float64// degree angle int with range [0,359] for player direction
	Face Pair[float64,float64]// movement direction with range [[-1,1],[-1,1]]
}

type PlayerComponent struct {
	Name string// username; ip for now
	Health int// current player health (cap enforced in update loop)
	Coins int// how much money the player has
	Weapon Weapon// current player weapon; default is 0 for Melee
	Loc Pair[float64, float64]// current location
	Dir Direction// direction player faces & direction player moves; currently, both are the same
	//MoveNum int// most recently-processed move
}

func (p PlayerComponent) String() string {
	s := ""
	s += "Name: " + p.Name + "\n"
	s += "Health: " + strconv.Itoa(p.Health) + "\n"
	s += "Coins: " + strconv.Itoa(p.Coins) + "\n"
	s += "Weapon: " + strconv.Itoa(int(p.Weapon)) + "\n"
	s += "Loc: " + strconv.FormatFloat(float64(p.Loc.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Loc.Second), 'e', -1, 32) + "\n"
	s += "Dir: " + strconv.FormatFloat(float64(p.Dir.Face.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Dir.Face.Second), 'e', -1, 32)

	return s
}

var (
	World			= inmem.NewECSWorld()
	ItemMapComp		= ecs.NewComponentType[ItemMapComponent]()// contains items
	PlayerMapComp	= ecs.NewComponentType[PlayerMapComponent]()// contains player locations
	PlayerComp		= ecs.NewComponentType[PlayerComponent]()
	ItemMap, PlayerMap storage.EntityID
	Players			= make(map[string] storage.EntityID, 0)//players are names and components identified by strings; input into a map to make it easier to add and remove components
	MoveTx			= ecs.NewTransactionType[Move]()//(World, "move")
)

const (
	tickRate		= 5// ticks per second
)

type Game struct {
	Dims	Pair[float64, float64]
	CSize	float64// cell size
	Players	[]string// list of players
}

var GameParams Game

type Move struct {
	Player	string
	Up		bool
	Down	bool
	Left	bool
	Right	bool
	//PacketNum int
}

type ModPlayer struct {// for adding and removing players
	Name	string
}

