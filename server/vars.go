package main

import (
	"strconv"
	"sync"
	"math"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type void struct{}
var pewp void

type Mult interface {
	getFirst() float64
	getSecond() float64
}

type Pair[T1 any, T2 any] struct {
	First T1
	Second T2
}

func (p Pair[float64, any]) getFirst() float64 {
	return p.First
}

func (p Pair[any, float64]) getSecond() float64 {
	return p.Second
}

type Triple[T1 any, T2 any, T3 any] struct {
	First T1
	Second T2
	Third T3
}

func (t Triple[float64, any, void]) getFirst() float64 {
	return t.First
}

func (t Triple[any, float64, void]) getSecond() float64 {
	return t.Second
}

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
	Dud = iota - 1// empty weapon
	Melee 
	Slug
)

type WeaponData struct {
	Attack int
	Range float64
}

type WeaponComponent struct {
	Loc Pair[float64, float64]
	Val Weapon// weapon type; TODO: implement ammo later outside of weapon component
	// cooldown, ammo, damage, range
}

type PlayerComponent struct {
	Name string// username; ip for now
	Health int// current player health (cap enforced in update loop)
	Coins int// how much money the player has
	Weapon Weapon// current player weapon; default is 0 for Melee
	Loc Pair[float64, float64]// current location
	Dir Pair[float64, float64]// array of movement directions with range [[-1,1],[-1,1]] where each pair is the movement at a given timestep (divided uniformly over the tick) and the first direction is the one that determines player movement
	Extract Pair[float64, float64]// extraction point; as long as the player is within some distance of the extraction point, player coins are offloaded
	IsRight bool// whether player is facing right
	MoveNum int// most recently-processed move
}

type BarePlayer struct {
	Name string
	Health int
	Coins int
	//Weapon Weapon
	//ExtractX float64
	//ExtractY float64
	LocX float64
	LocY float64
	IsRight bool
	InputNum int
}

type TestPlayer struct {
	Name string
	Health int
	Coins int
	Weapon Weapon
	ExtractX float64
	ExtractY float64
	LocX float64
	LocY float64
}

func (p PlayerComponent) Simplify() BarePlayer {
	return BarePlayer{p.Name, p.Health, p.Coins, p.Loc.First, p.Loc.Second, p.IsRight, p.MoveNum}// update Simplify for weapons & extraction point
}

func (p PlayerComponent) Testify() TestPlayer {
	return TestPlayer{p.Name, p.Health, p.Coins, p.Weapon, p.Extract.First, p.Extract.Second, p.Loc.First, p.Loc.Second}
}

func (p PlayerComponent) String() string {
	s := ""
	s += "Name: " + p.Name + "\n"
	s += "Health: " + strconv.Itoa(p.Health) + "\n"
	s += "Coins: " + strconv.Itoa(p.Coins) + "\n"
	s += "Weapon: " + strconv.Itoa(int(p.Weapon)) + "\n"
	s += "Loc: " + strconv.FormatFloat(float64(p.Loc.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Loc.Second), 'e', -1, 32) + "\n"
	s += "Dir: " + strconv.FormatFloat(float64(p.Dir.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Dir.Second), 'e', -1, 32)

	return s
}

var (
	World					= inmem.NewECSWorld()
	CoinMap					= make(map[Pair[int, int]] map[Pair[storage.EntityID, Pair[float64,float64]]] void)// maps cells to sets of coin lists
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
	DefaultWeapon Weapon	= Dud
	Attacks					= make([]AttackTriple, 0)
)

const (
	TickRate			= 5// ticks per second
	ClientTickRate		= 60// used to determine tickrate relative to cardinal server
	PlayerRadius		= 0.5// used to determine which coins to collect
	ExtractionRadius	= 10// determines when players are in range of their extraction point
	sped				= 2// player speed
)

type Game struct {
	Dims	Pair[float64, float64]
	CSize	float64// cell size
	Players	[]string// list of players
}

var GameParams Game

type Move struct {
	PlayerID				string
	Up						bool
	Down					bool
	Left					bool
	Right					bool
	Input_sequence_number	int
	Delta					float64
}

type AttackTriple struct {
	AttackerID, DefenderID	string
	Damage					int
}

type AddPlayer struct {// for adding and removing players
	Name	string
	Coins	int
}

type ModPlayer struct {// for adding and removing players
	Name	string
}

func GetCell(loc Mult) Pair[int,int] {
	return Pair[int,int]{int(math.Floor(loc.getFirst()/GameParams.CSize)), int(math.Floor(loc.getSecond()/GameParams.CSize))}
}
