package main

import (
	"fmt"
	//"net/http"
	//"time"
	"strconv"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	//"github.com/argus-labs/world-engine/cardinal/ecs/component"
)

type void struct{}
var pewp void

type Pair[T1 any, T2 any] struct {
	First T1
	Second T2
}

type ItemComponent struct {}

type HealthComponent struct {
	Loc Pair[int, int]// location
	Val int// how much health it contains
}

type CoinComponent struct {
	Loc Pair[int, int]
	Val int// how many coins the component represents; could represent different denominations as larger or different colored coins
}

type Weapon int

const (// add more weapons as needed
	Melee = iota
	Sniper
)

type WeaponComponent struct {
	Loc Pair[int, int]
	Val Weapon// weapon type; TODO: implement ammo later outside of weapon component
	// cooldown, ammo, damage, range
}

type ItemMapComponent struct {// TODO: maybe don't turn these into components
	Items map[Pair[int, int]] map[Pair[storage.EntityID, Pair[int,int]]] void// maps cells to sets of item lists
}

type PlayerMapComponent struct {
	Players map[Pair[int,int]] map[Pair[storage.EntityID, Pair[int,int]]] void// maps cells to sets of player name-location pairs
}

type Direction struct {
	Theta int// degree angle int with range [0,359] for player direction
	Face Pair[int,int]// movement direction with range [[-1,1],[-1,1]]
}

type PlayerComponent struct {
	Name string// username; ip for now
	Health int// current player health (cap enforced in update loop)
	Coins int// how much money the player has
	Weapon Weapon// current player weapon; default is 0 for Melee
	Loc Pair[int, int]// current location
	Dir Direction// direction player faces & direction player moves; currently, both are the same
}

func (p PlayerComponent) String() string {
	s := ""
	s += "Name: " + p.Name + "\n"
	s += "Health: " + strconv.Itoa(p.Health) + "\n"
	s += "Coins: " + strconv.Itoa(p.Coins) + "\n"
	s += "Weapon: " + strconv.Itoa(int(p.Weapon)) + "\n"
	s += "Loc: " + strconv.Itoa(p.Loc.First) + " " + strconv.Itoa(p.Loc.Second) + "\n"
	s += "Dir: " + strconv.Itoa(p.Dir.Face.First) + " " + strconv.Itoa(p.Dir.Face.Second)

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
	Dims	Pair[int, int]
	CSize	int// cell size
	Players	[]string// list of players
}

var GameParams Game

type Move struct {
	Player	string
	Up		bool
	Down	bool
	Left	bool
	Right	bool
}

type ModPlayer struct {// for adding and removing players
	Name	string
}

func init() {
	World.RegisterComponents(ItemMapComp, PlayerMapComp, PlayerComp)
	MoveTx.SetID(0)
	//World.RegisterTransactions(MoveTx)
}

//func callback(cb func() error) {
//	cb()
//	time.Sleep(time.Second/tickRate)
//	callback(cb)
//}

// adds update Systems
func processMoves(World *ecs.World, q *ecs.TransactionQueue) error {// adjusts player directions based on their movement
	moveMap := make(map[string] Move)

	for _, move := range MoveTx.In(q) {
		moveMap[move.Player] = move
	}

	for name, move := range moveMap {
		PlayerComp.Update(World, Players[name], func(comp PlayerComponent) PlayerComponent {// modifies player direction struct
			diff := func(a, b bool) int {
				if a == b { return 0 }
				if a && !b { return 1 }
				return -1
			}

			comp.Dir.Face = Pair[int,int]{diff(move.Right, move.Left), diff(move.Up, move.Down)}// adjusts move direction
			return comp
		})
	}

	return nil
}

func makeMoves(World *ecs.World, q *ecs.TransactionQueue) error {// moves player based on the coin-speed; TODO: check collisions
	for playerName, id := range Players {
		tmpPlayer, err := PlayerComp.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		loc := Pair[int,int]{prevLoc.First + (10 * tmpPlayer.Dir.Face.First)/(1 + tmpPlayer.Coins), prevLoc.Second + (10 * tmpPlayer.Dir.Face.Second)/(1 + tmpPlayer.Coins)}// change speed function

		PlayerComp.Update(World, Players[playerName], func(comp PlayerComponent) PlayerComponent{// modifies player location
			comp.Loc = loc
			return comp
		})

		PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent{// moves player in map
			delete(comp.Players[Pair[int,int]{prevLoc.First/GameParams.CSize, prevLoc.Second/GameParams.CSize}], Pair[storage.EntityID, Pair[int,int]]{id, prevLoc})
			comp.Players[Pair[int,int]{loc.First/GameParams.CSize, loc.Second/GameParams.CSize}][Pair[storage.EntityID,Pair[int,int]]{id, loc}] = pewp

			return comp
		})
	}
	return nil
}

func HandleCreateGame(game Game) error {
	GameParams = game
	ItemMap, err := World.Create(ItemMapComp)// creates an ItemMap entity
	PlayerMap, err := World.Create(PlayerMapComp)// creates a PlayerMap entity
	playerIDs, err := World.CreateMany(len(GameParams.Players), PlayerComp)// creates player entities

	for i, playername := range GameParams.Players {// associates storage.EntityIDs with each player
		Players[playername] = playerIDs[i]
	}

	if err != nil {
		return fmt.Errorf("Error initializing game objects: %w", err)
	}

	// initializes player and item maps
	itemmap := make(map[Pair[int, int]] map[Pair[storage.EntityID, Pair[int,int]]] void)
	playermap := make(map[Pair[int, int]] map[Pair[storage.EntityID, Pair[int, int]]] void)
	for i := 0; i <= GameParams.Dims.First/GameParams.CSize; i++ {
		for j := 0; j <= GameParams.Dims.Second/GameParams.CSize; j++ {
			itemmap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[int,int]]] void)
			playermap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[int, int]]] void)
		}
	}

	ItemMapComp.Set(World, ItemMap, ItemMapComponent{itemmap})// initializes ItemMap using empty map
	PlayerMapComp.Set(World, ItemMap, PlayerMapComponent{playermap})// initializes PlayerMap using empty map

	for _, playername := range GameParams.Players {
		PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Melee, Pair[int,int]{25,25}, Direction{90, Pair[int,int]{0,0}}})// initializes player entitities through their component

		PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent {// adds players to the board
			playercomp, err := PlayerComp.Get(World, Players[playername])

			if err != nil {
				fmt.Errorf("Error getting location with callback function: %w", err)
				return comp
			}

			newPlayer := Pair[storage.EntityID, Pair[int,int]]{Players[playername], playercomp.Loc}
			comp.Players[Pair[int,int]{25/GameParams.CSize,25/GameParams.CSize}][newPlayer] = pewp

			return comp
		})
	}

	World.AddSystem(processMoves)
	World.AddSystem(makeMoves)
	World.LoadGameState()

	// calls callback goroutine to keep World ticking

	//go func(){ TODO enable after testing
	//	for range time.Tick(time.Second/tickRate) {
	//		World.Tick()
	//	}
	//}()

	return nil
}

func HandlePlayerPush(player ModPlayer) error {
	playerID, err := World.Create(PlayerComp)// creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Melee, Pair[int,int]{25,25}, Direction{90, Pair[int,int]{0,0}}})// default player
	PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent {// adds a player to the board
		playercomp, err := PlayerComp.Get(World, Players[player.Name])

		if err != nil {
			fmt.Errorf("Error getting location with callback function: %w", err)
			return comp
		}

		newPlayer := Pair[storage.EntityID, Pair[int,int]]{Players[player.Name], playercomp.Loc}
		comp.Players[Pair[int,int]{25/GameParams.CSize,25/GameParams.CSize}][newPlayer] = pewp

		return comp
	})

	return nil
}

func HandlePlayerPop(player ModPlayer) error {
	//err := World.Remove(Players[player.Name])

	//if err != nil {
	//	return err
	//}

	PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent{//removes a player from the board
		playercomp, err := PlayerComp.Get(World, Players[player.Name])

		if err != nil {
			fmt.Errorf("error removing player: %w", err)
			return comp
		}

		oldPlayer := Pair[storage.EntityID, Pair[int,int]]{Players[player.Name], playercomp.Loc}
		delete(comp.Players[Pair[int,int]{playercomp.Loc.First/GameParams.CSize, playercomp.Loc.Second/GameParams.CSize}], oldPlayer)

		return comp
	})

	delete(Players, player.Name)

	return nil
}

func TickTock() {// testing function used to make the game tick
	World.Tick()
}

func GetPlayerState(player ModPlayer) (PlayerComponent, error) {// testing function used in place of broadcast to get state of players
	comp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		return PlayerComponent{}, fmt.Errorf("Player fetch error: %w", err)
	}

	return comp, nil
}

func HandleMakeMove(move Move) {
	MoveTx.AddToQueue(World, move)// adds "move" transaction to World transaction queue
}
