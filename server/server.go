package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

// world Systems
func processMoves(World *ecs.World, q *ecs.TransactionQueue) error {// adjusts player directions based on their movement
	moveMap := make(map[string] Move)

	for _, move := range MoveTx.In(q) {
		_, contains := moveMap[move.PlayerID]
		if !contains || moveMap[move.PlayerID].PacketNum <= move.PacketNum{
			moveMap[move.PlayerID] = move
		}
	}

	for name, move := range moveMap {
		naMe, contains := Players[name]
		
		if !contains {
			return fmt.Errorf("Cardinal: unregistered player attempting to move")
		}

		PlayerComp.Update(World, naMe, func(comp PlayerComponent) PlayerComponent {// modifies player direction struct
			diff := func(a, b bool) float64 {
				if a == b { return 0 }
				if a && !b { return 1 }
				return -1
			}

			comp.Dir.Face = Pair[float64,float64]{diff(move.Right, move.Left), diff(move.Up, move.Down)}// adjusts move direction
			comp.MoveNum = move.PacketNum

			return comp
		})
	}

	return nil
}

func makeMoves(World *ecs.World, q *ecs.TransactionQueue) error {// moves player based on the coin-speed
	for playerName, id := range Players {
		tmpPlayer, err := PlayerComp.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		bound := func(x float64, y float64) (float64, float64){
			return math.Min(float64(GameParams.Dims.First), math.Max(0, x)), math.Min(float64(GameParams.Dims.Second), math.Max(0, y))
		}

		x, y := bound(prevLoc.First + (10 * tmpPlayer.Dir.Face.First)/float64(1 + tmpPlayer.Coins), prevLoc.Second + (10 * tmpPlayer.Dir.Face.Second)/float64(1 + tmpPlayer.Coins))
		
		loc := Pair[float64, float64]{x,y}// change speed function

		PlayerComp.Update(World, Players[playerName], func(comp PlayerComponent) PlayerComponent{// modifies player location
			comp.Loc = loc
			return comp
		})

		delete(PlayerMap[Pair[int,int]{int(math.Floor(prevLoc.First/GameParams.CSize)), int(math.Floor(prevLoc.Second/GameParams.CSize))}], Pair[storage.EntityID, Pair[float64,float64]]{id, prevLoc})
		PlayerMap[Pair[int,int]{int(math.Floor(loc.First/GameParams.CSize)), int(math.Floor(loc.Second/GameParams.CSize))}][Pair[storage.EntityID,Pair[float64,float64]]{id, loc}] = pewp
	}
	return nil
}


func HandlePlayerPush(player ModPlayer) error {
	if _, contains := Players[player.Name]; contains {// player already exists; don't do anything
		return nil
	}

	playerID, err := World.Create(PlayerComp)// creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Melee, Pair[float64,float64]{25,25}, Direction{90, Pair[float64,float64]{0,0}}, 0})// default player

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting location with callback function: %w", err)
	}

	newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
	PlayerMap[Pair[int,int]{25/int(GameParams.CSize),25/int(GameParams.CSize)}][newPlayer] = pewp

	return nil
}

func HandlePlayerPop(player ModPlayer) error {
	//err := World.Remove(Players[player.Name])

	//if err != nil {
	//	return err
	//}

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("error removing player: %w", err)
	}

	oldPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
	delete(PlayerMap[Pair[int,int]{int(math.Floor(playercomp.Loc.First/GameParams.CSize)), int(math.Floor(playercomp.Loc.Second/GameParams.CSize))}], oldPlayer)

	delete(Players, player.Name)

	return nil
}

func TickTock() error {// testing function used to make the game tick
	return World.Tick()
}

func GetPlayerState(player ModPlayer) (PlayerComponent, error) {// testing function used in place of broadcast to get state of players
	if _, contains := Players[player.Name]; contains == false {
		return PlayerComponent{}, fmt.Errorf("Player does not exist")
	}

	comp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		return PlayerComponent{}, fmt.Errorf("Player fetch error: %w", err)
	}

	return comp, nil
}

func HandleMakeMove(move Move) {
	MoveTx.AddToQueue(World, move)// adds "move" transaction to World transaction queue
}

func CreateGame(game Game) error {
	//if World.stateIsLoaded {
	//	return fmt.Errorf("already loaded state")
	//}
	GameParams = game
	World.RegisterComponents(PlayerComp, CoinComp, HealthComp, WeaponComp)
	World.AddSystem(processMoves)
	World.AddSystem(makeMoves)

	World.LoadGameState()
	MoveTx.SetID(0)
	//ItemMap, err := World.Create(ItemMapComp)// creates an ItemMap entity
	//PlayerMap, err := World.Create(PlayerMapComp)// creates a PlayerMap entity
	playerIDs, err := World.CreateMany(len(GameParams.Players), PlayerComp)// creates player entities

	Players = make(map[string] storage.EntityID)
	for i, playername := range GameParams.Players {// associates storage.EntityIDs with each player
		Players[playername] = playerIDs[i]
	}

	if err != nil {
		return fmt.Errorf("Error initializing game objects: %w", err)
	}

	Width = int(math.Ceil(GameParams.Dims.First/GameParams.CSize))
	Height = int(math.Ceil(GameParams.Dims.Second/GameParams.CSize))

	// initializes player and item maps
	for i := 0; i <= Width; i++ {
		for j := 0; j <= Height; j++ {
			CoinMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64,float64]]] void)
			HealthMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64,float64]]] void)
			WeaponMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64,float64]]] void)
			PlayerMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64, float64]]] void)
		}
	}

	for _, playername := range GameParams.Players {
		PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Melee, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Direction{90, Pair[float64,float64]{0,0}}, 0})// initializes player entitities through their component

		playercomp, err := PlayerComp.Get(World, Players[playername])

		if err != nil {
			fmt.Errorf("Error getting location with callback function: %w", err)
		}

		newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[playername], playercomp.Loc}
		PlayerMap[Pair[int,int]{25/int(GameParams.CSize),25/int(GameParams.CSize)}][newPlayer] = pewp
	}

	return nil
}

func distance(loc1, loc2 Pair[float64, float64]) float64 {// returns distance between two coins
	return math.Sqrt(math.Pow(loc1.First - loc2.First, 2) + math.Pow(loc1.Second - loc2.Second, 2))
}

func SpawnCoins() error {// randomly spawn 5 coins in each cell and don't place if a coin exists nearby
	var (
		coinCellNum = 5
		coinRadius = 0.1// <= GameParams.CSize/2
		density = 0.01 // number of coins per square unit
		maxCoinsInCell = int(math.Pow(GameParams.CSize, 2)*density)
	)

	newCoins := make([]Pair[float64, float64], 0)
	deleteList := make([]Pair[float64, float64], 0)
	mutex := &sync.RWMutex{}

	mutex.RLock()
	for i := 0; i < Width; i++ {
		for j := 0; j < Height; j++ {
			if len(CoinMap[Pair[int,int]{i,j}]) >= maxCoinsInCell { continue }

			coinSet := make(map[Pair[float64,float64]] void)// making coins near but not at the edge is fine because the expected distance between coins on opposite sides of the edge will be double the expected distance between coins on the same side
			for k := 0; k < coinCellNum; k++ {
				coinSet[Pair[float64,float64]{float64(i)*GameParams.CSize + coinRadius + rand.Float64()*(GameParams.CSize-2*coinRadius), float64(j)*GameParams.CSize + coinRadius + rand.Float64()*(GameParams.CSize-2*coinRadius)}] = pewp
			}

			if len(deleteList) > 0 {
				deleteList = make([]Pair[float64, float64], 0)
			}
			for coin,_ := range CoinMap[Pair[int,int]{i, j}] {// concurrent iteration and write
				for coinPos, _ := range coinSet {
					if distance(coinPos, coin.Second) <= coinRadius {
						deleteList = append(deleteList, coinPos)
					}
				}
			}

			for _, coinPos := range deleteList {
				delete(coinSet, coinPos)
			}

			for coin, _ := range coinSet {
				newCoins = append(newCoins, coin)
			}

		}
	}
	mutex.RUnlock()

	//create mutex to prevent concurrent ticks from causing problems; iterating through map above takes too much time to do, so when the second tick is called and iteration occurs, the first tick is still trying to add elements to the map
	// also limit the number of coins in each cell of the coinmap and the size of the map so we don't have iteration problems

	mutex.Lock()
	for _, coin := range newCoins {
		coinID, err := World.Create(CoinComp)

		if err != nil {
			return fmt.Errorf("Coin creation failed: %w", err)
		}

		CoinMap[Pair[int,int]{int(math.Floor(coin.First/GameParams.CSize)), int(math.Floor(coin.Second/GameParams.CSize))}][Pair[storage.EntityID, Pair[float64, float64]]{coinID, coin}] = pewp
		CoinComp.Set(World, coinID, CoinComponent{coin, 1})
	}
	mutex.Unlock()

	return nil
}
