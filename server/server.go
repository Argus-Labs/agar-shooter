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
	moveMap := make(map[string] []Move)

	for _, move := range MoveTx.In(q) {
		if _, contains := moveMap[move.PlayerID]; !contains {
			/*
			pcomp, err := PlayerComp.Get(World, Players[move.PlayerID])

			if err != nil {
				return err
			}
			
			if pcomp.MoveNum != move.Input_sequence_number - 1{
				fmt.Printf("Difference in input sequence number is not 1; received sequence number %i after sequence number %i.",move.Input_sequence_number,pcomp.MoveNum)
				return nil
			}
			*/

			moveMap[move.PlayerID] = []Move{move}
		} else {
			/*
			if num := moveMap[move.PlayerID][len(moveMap[move.PlayerID])-1].Input_sequence_number;move.Input_sequence_number != num + 1 {
				fmt.Printf("Difference in input sequence number is not 1; received sequence number %i after sequence number %i.",move.Input_sequence_number,num)
				return nil
			}
			*/
			moveMap[move.PlayerID] = append(moveMap[move.PlayerID], move)
		}
	}

	for name, moveList := range moveMap {
		entityID, contains := Players[name]
	
		if !contains {
			str := ""

			for key, _ := range Players{
				str += " " + key
			}

			return fmt.Errorf("Cardinal: unregistered player attempting to move " + str)
		}

		var dir Pair[float64, float64]

		diff := func(a, b bool) float64 {
			if a == b { return 0 }
			if a && !b { return 1 }
			return -1
		}

		for _, move := range moveList {
			moove := Pair[float64,float64]{diff(move.Right, move.Left), diff(move.Up, move.Down)}
			norm := math.Max(1, math.Sqrt(moove.First*moove.First + moove.Second*moove.Second))

			dir = Pair[float64, float64]{dir.First + move.Delta*moove.First/norm, dir.Second + move.Delta*moove.Second/norm}
		}

		lastMove := Pair[float64,float64]{diff(moveList[len(moveList)-1].Right, moveList[len(moveList)-1].Left), diff(moveList[len(moveList)-1].Up, moveList[len(moveList)-1].Down)}

		PlayerComp.Update(World, entityID, func(comp PlayerComponent) PlayerComponent {// modifies player direction struct
			comp.Dir = dir// adjusts move directions
			comp.MoveNum = moveList[len(moveList)-1].Input_sequence_number
			if lastMove.First != 0 {
				comp.IsRight = lastMove.First > 0
			}

			return comp
		})
	}

	return nil
}

func bound(x float64, y float64) Pair[float64, float64]{
	return Pair[float64, float64]{math.Min(float64(GameParams.Dims.First), math.Max(0, x)), math.Min(float64(GameParams.Dims.Second), math.Max(0, y))}
}

func distance(loc1, loc2 Pair[float64, float64]) float64 {// returns distance between two coins
	return math.Sqrt(math.Pow(loc1.First - loc2.First, 2) + math.Pow(loc1.Second - loc2.Second, 2))
}

func move(tmpPlayer PlayerComponent) Pair[float64, float64] {// change speed function
	dir := tmpPlayer.Dir
	coins := 0//tmpPlayer.Coins
	return bound(tmpPlayer.Loc.First + (sped * dir.First)/(float64(1 + coins)), tmpPlayer.Loc.Second + (sped * dir.Second)/(float64(1 + coins)))
}

func CoinProjDist(start, end, coin Pair[float64, float64]) float64 {// closest distance the coin is from the player obtained by checking the orthogonal projection of the coin with the segment defined by [start,end] TODO: write testcase for finding this value
	vec := Pair[float64, float64]{end.First-start.First, end.Second-start.Second}
	coeff := (vec.First*coin.First + vec.Second*coin.Second)/(vec.First*vec.First + vec.Second*vec.Second)
	proj := Pair[float64, float64]{coeff*vec.First + start.First, coeff*vec.Second + start.Second}
	ortho := Pair[float64, float64]{coin.First - proj.First, coin.Second-proj.Second}

	if proj.First*vec.First + proj.Second*vec.Second < 0 {// if the coin is outside of the span of the orthogonal, return the distance to the closest endpoint
		return math.Min(math.Sqrt(math.Pow(coin.First - start.First, 2) + math.Pow(coin.Second - start.Second, 2)), math.Sqrt(math.Pow(coin.First - end.First, 2) + math.Pow(coin.Second - end.Second, 2)))
	}

	return math.Sqrt(ortho.First*ortho.First + ortho.Second*ortho.Second)
}

func attack(id storage.EntityID, weapon Weapon) error {// attack a player; TODO: change attacking to be based on IsRight
	kill := false
	var name string
	PlayerComp.Update(World, id, func(comp PlayerComponent) PlayerComponent{// modifies player location
		comp.Health -= Weapons[weapon].Attack
		kill = comp.Health <= 0
		name = comp.Name

		return comp
	})

	if kill {// removes player from map if they die
		if err := HandlePlayerPop(ModPlayer{name}); err != nil {
			return err
		}
	}

	return nil
}

func makeMoves(World *ecs.World, q *ecs.TransactionQueue) error {// moves player based on the coin-speed
	attackQueue := make([]Pair[storage.EntityID, Weapon],0)
	for playerName, id := range Players {
		tmpPlayer, err := PlayerComp.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		// attacking players; each player attacks the closest player TODO: change targetting system later

		var (
			minID storage.EntityID
			minDistance float64
		)

		assigned := false

		for closestPlayer, _ := range PlayerMap[Pair[int,int]{int(math.Floor(prevLoc.First/GameParams.Dims.First)),int(math.Floor(prevLoc.Second/GameParams.Dims.Second))}] {
			dist := distance(closestPlayer.Second, prevLoc)
			if closestPlayer.First != id {
				if !assigned {
					minID = closestPlayer.First
					minDistance = dist
					assigned = true
				} else if minDistance > dist {
					minID = closestPlayer.First
					minDistance = dist
				}
			}
		}

		if assigned && minDistance <= Weapons[tmpPlayer.Weapon].Range {
			attackQueue = append(attackQueue, Pair[storage.EntityID, Weapon]{minID, tmpPlayer.Weapon})
		}

		// moving players

		loc := move(tmpPlayer)

		delete(PlayerMap[Pair[int,int]{int(math.Floor(prevLoc.First/GameParams.CSize)), int(math.Floor(prevLoc.Second/GameParams.CSize))}], Pair[storage.EntityID, Pair[float64,float64]]{id, prevLoc})
		PlayerMap[Pair[int,int]{int(math.Floor(loc.First/GameParams.CSize)), int(math.Floor(loc.Second/GameParams.CSize))}][Pair[storage.EntityID,Pair[float64,float64]]{id, loc}] = pewp
		
		hitCoins := make([]Pair[storage.EntityID, Pair[float64,float64]], 0)

		for i := int(math.Floor(prevLoc.First/GameParams.CSize)); i <= int(math.Floor(loc.First/GameParams.CSize)); i++ {
			for j := int(math.Floor(prevLoc.Second/GameParams.CSize)); j <= int(math.Floor(loc.Second/GameParams.CSize)); j++ {
				for coin, _ := range CoinMap[Pair[int, int]{i,j}] {
					if distance(prevLoc, coin.Second) <= PlayerRadius || distance(loc, coin.Second) <= PlayerRadius {//CoinProjDist(prevLoc, loc, coin.Second) <= PlayerRadius {
						hitCoins = append(hitCoins, coin)
					}
				}
			}
		}

		extraCoins := 0

		for _, entityID := range hitCoins {
			extraCoins++// TODO: change this to the actual coin value
			delete(CoinMap[Pair[int,int]{int(math.Floor(entityID.Second.First/GameParams.CSize)),int(math.Floor(entityID.Second.Second/GameParams.CSize))}], entityID)

			if err := World.Remove(entityID.First); err != nil {
				return err
			}
		}

		PlayerComp.Update(World, Players[playerName], func(comp PlayerComponent) PlayerComponent{// modifies player location
			comp.Loc = loc
			comp.Coins += extraCoins
			
			if distance(comp.Loc, comp.Extract) <= ExtractionRadius {// extraction point offloading
				comp.Coins = 0
			}

			return comp
		})
	}

	for _, pair := range attackQueue {
		if err := attack(pair.First, pair.Second); err != nil {
			return err
		}
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

	PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Dud, Pair[float64,float64]{25 + (rand.Float64()-0.5)*10,25 + (rand.Float64()-0.5)*10}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, true, -1})// default player
	//PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// default player

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting location with callback function: %w", err)
	}

	newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
	PlayerMap[Pair[int,int]{25/int(GameParams.CSize),25/int(GameParams.CSize)}][newPlayer] = pewp

	return nil
}

func HandlePlayerPop(player ModPlayer) error {
	err := World.Remove(Players[player.Name])

	if err != nil {
		return err
	}

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
	err := World.Tick()
	return err
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

func GetPlayerStatus() ([]Pair[string, Pair[float64, float64]]) {// sends all player information to each player
	locs := make([]Pair[string, Pair[float64, float64]], 0)
	for key, id := range Players {
		comp, _ := PlayerComp.Get(World, id)
		locs = append(locs, Pair[string, Pair[float64, float64]]{key, comp.Loc})
	}

	return locs
}

func HandleMakeMove(move Move) {
	MoveTx.AddToQueue(World, move)// adds "move" transaction to World transaction queue
}

func CreateGame(game Game) error {
	//if World.stateIsLoaded {
	//	return fmt.Errorf("already loaded state")
	//}
	if game.CSize == 0 {
		return fmt.Errorf("Cardinal: cellsize is zero")
	}
	GameParams = game
	World.RegisterComponents(PlayerComp, CoinComp, HealthComp, WeaponComp)
	World.AddSystem(processMoves)
	World.AddSystem(makeMoves)

	World.LoadGameState()
	MoveTx.SetID(0)
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
		PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Dud, Pair[float64,float64]{25 + (rand.Float64()-0.5)*10,25 + (rand.Float64()-0.5)*10}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, true, -1})// initializes player entitities through their component
		//PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// initializes player entitities through their component

		playercomp, err := PlayerComp.Get(World, Players[playername])

		if err != nil {
			fmt.Errorf("Cardinal: Error getting location with callback function: %w", err)
		}

		newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[playername], playercomp.Loc}
		PlayerMap[Pair[int,int]{25/int(GameParams.CSize),25/int(GameParams.CSize)}][newPlayer] = pewp
	}

	return nil
}

func SpawnCoins(mutex *sync.RWMutex) error {// randomly spawn 5 coins in each cell and don't place if a coin exists nearby
	var (
		coinCellNum = 1
		coinRadius = 0.5// <= GameParams.CSize/2
		density = 0.1// number of coins per square unit
		maxCoinsInCell = int(math.Ceil(math.Pow(GameParams.CSize, 2)*density))
	)

	newCoins := make([]Pair[float64, float64], 0)

	mutex.RLock()
	for i := 0; i < Width; i++ {
		for j := 0; j < Height; j++ {
			if len(CoinMap[Pair[int,int]{i,j}]) >= maxCoinsInCell { continue }

			for k := 0; k < coinCellNum; k++ {
				newCoin := Pair[float64,float64]{float64(i)*GameParams.CSize + coinRadius + rand.Float64()*(GameParams.CSize-2*coinRadius), float64(j)*GameParams.CSize + coinRadius + rand.Float64()*(GameParams.CSize-2*coinRadius)}
				keep := true

				for coin,_ := range CoinMap[Pair[int,int]{i, j}] {// concurrent iteration and write
					keep = keep && (distance(coin.Second, newCoin) <= coinRadius)
				}

				for player, _ := range PlayerMap[Pair[int,int]{i,j}] {
					keep = keep && (distance(player.Second, newCoin) <= PlayerRadius+1)
				}

				if keep {
					newCoins = append(newCoins, newCoin)
				}
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

func NearbyCoins(player ModPlayer) Pair[[]float64, []float64] {
	xloc := make([]float64, 0)
	yloc := make([]float64, 0)
	
	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting location with callback function: %w", err)
	}

	for i := math.Max(0, math.Floor((playercomp.Loc.First-ClientView.First)/GameParams.CSize)); i <= math.Min(float64(Width), math.Ceil((playercomp.Loc.First+ClientView.First)/GameParams.CSize)); i++ {
		for j := math.Max(0, math.Floor((playercomp.Loc.Second-ClientView.Second)/GameParams.CSize)); j <= math.Min(float64(Height), math.Ceil((playercomp.Loc.Second+ClientView.Second)/GameParams.CSize)); j++ {
			for coin, _ := range CoinMap[Pair[int,int]{int(i),int(j)}] {
				xloc = append(xloc, coin.Second.First)
				yloc = append(yloc, coin.Second.Second)
			}
		}
	}

	return Pair[[]float64, []float64]{xloc, yloc}
}

func AddTestPlayer(player PlayerComponent) error {
	if _, contains := Players[player.Name]; contains {// player already exists; don't do anything
		return nil
	}

	playerID, err := World.Create(PlayerComp)// creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], player)// default player

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		return fmt.Errorf("Error getting location with callback function: %w", err)
	}

	newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
	PlayerMap[Pair[int,int]{int(math.Floor(player.Loc.First/GameParams.CSize)),int(math.Floor(player.Loc.Second/GameParams.CSize))}][newPlayer] = pewp

	return nil
}
