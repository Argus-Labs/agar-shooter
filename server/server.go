package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

func HandlePlayerPush(player AddPlayer) error {
	if _, contains := Players[player.Name]; contains {// player already exists; don't do anything
		fmt.Println("Player already exists; not pushing again")
		return nil
	}

	playerID, err := World.Create(PlayerComp)// creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, player.Coins, DefaultWeapon, Pair[float64,float64]{25 + (rand.Float64()-0.5)*10,25 + (rand.Float64()-0.5)*10}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, true, -1})// default player
	//PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// default player

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting location with callback function: %w", err)
	}

	newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
	PlayerMap[GetCell(playercomp.Loc)][newPlayer] = pewp

	return nil
}

func HandlePlayerPop(player ModPlayer) error {
	playercomp, err := PlayerComp.Get(World, Players[player.Name])
	if err != nil {
		return err
	}
	
	if err = World.Remove(Players[player.Name]); err != nil {
		fmt.Errorf("error removing player: %w", err)
	}

	// put all coins around the player
	coins := playercomp.Coins
	tot := int(math.Max(1, float64(coins/10 + (coins%10)/5 + coins%5)))
	start := 0
	rad := float64(tot)/(2*math.Pi)
	newCoins := make([]Triple[float64, float64, int], 0)

	for coins > 0 {// decomposes into 10s, 5s, 1s
		addCoins := 0
		switch {
			case coins >= 10: {
				addCoins = 10
				coins -= 10
				break
			}
			case coins >= 5: {
				addCoins = 5
				coins -= 5
				break
			}
			default: {
				addCoins = 1
				coins--
			}
		}

		peep := bound(playercomp.Loc.First + rad*math.Cos(2*math.Pi*float64(start)/float64(tot)), playercomp.Loc.Second + rad*math.Sin(2*math.Pi*float64(start)/float64(tot)))
		newCoins = append(newCoins, Triple[float64, float64, int]{peep.First, peep.Second, addCoins})
		start++
	}

	for _, coin := range newCoins {
		coinID, err := World.Create(CoinComp)

		if err != nil {
			return fmt.Errorf("Coin creation failed: %w", err)
		}

		CoinMap[GetCell(coin)][Pair[storage.EntityID, Triple[float64, float64, int]]{coinID, Triple[float64, float64, int]{coin.First, coin.Second, coin.Third}}] = pewp
		CoinComp.Set(World, coinID, CoinComponent{Pair[float64, float64]{coin.First, coin.Second}, coin.Third})
	}
	

	oldPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
	delete(PlayerMap[GetCell(playercomp.Loc)], oldPlayer)

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
			CoinMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Triple[float64,float64,int]]] void)
			HealthMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64,float64]]] void)
			WeaponMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64,float64]]] void)
			PlayerMap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64, float64]]] void)
		}
	}

	for _, playername := range GameParams.Players {
		PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, DefaultWeapon, Pair[float64,float64]{25 + (rand.Float64()-0.5)*10,25 + (rand.Float64()-0.5)*10}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, true, -1})// initializes player entitities through their component
		//PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// initializes player entitities through their component

		playercomp, err := PlayerComp.Get(World, Players[playername])

		if err != nil {
			fmt.Errorf("Cardinal: Error getting location with callback function: %w", err)
		}

		newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[playername], playercomp.Loc}
		PlayerMap[GetCell(playercomp.Loc)][newPlayer] = pewp
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

	newCoins := make([]Triple[float64, float64, int], 0)

	mutex.RLock()
	for i := 0; i < Width; i++ {
		for j := 0; j < Height; j++ {
			if len(CoinMap[Pair[int,int]{i,j}]) >= maxCoinsInCell { continue }

			for k := 0; k < coinCellNum; k++ {
				newCoin := Triple[float64,float64,int]{float64(i)*GameParams.CSize + coinRadius + rand.Float64()*(GameParams.CSize-2*coinRadius), float64(j)*GameParams.CSize + coinRadius + rand.Float64()*(GameParams.CSize-2*coinRadius), 1}
				keep := true

				for coin,_ := range CoinMap[Pair[int,int]{i, j}] {// concurrent iteration and write
					keep = keep && (distance(coin.Second, newCoin) > coinRadius)
				}

				for player, _ := range PlayerMap[Pair[int,int]{i,j}] {
					keep = keep && (distance(player.Second, newCoin) > PlayerRadius+1)
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
	// maybe make this a system so it can be run async

	mutex.Lock()
	for _, coin := range newCoins {
		coinID, err := World.Create(CoinComp)

		if err != nil {
			return fmt.Errorf("Coin creation failed: %w", err)
		}

		CoinMap[Pair[int,int]{int(math.Floor(coin.First/GameParams.CSize)), int(math.Floor(coin.Second/GameParams.CSize))}][Pair[storage.EntityID, Triple[float64, float64, int]]{coinID, coin}] = pewp
		CoinComp.Set(World, coinID, CoinComponent{Pair[float64,float64]{coin.First,coin.Second}, 1})
	}
	mutex.Unlock()

	return nil
}

func NearbyCoins(player ModPlayer) []NearbyCoin {
	coins := make([]NearbyCoin, 0)

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting player component: %w", err)
	}

	for i := math.Max(0, math.Floor((playercomp.Loc.First-ClientView.First/2)/GameParams.CSize)); i <= math.Min(float64(Width), math.Ceil((playercomp.Loc.First+ClientView.First/2)/GameParams.CSize)); i++ {
		for j := math.Max(0, math.Floor((playercomp.Loc.Second-ClientView.Second/2)/GameParams.CSize)); j <= math.Min(float64(Height), math.Ceil((playercomp.Loc.Second+ClientView.Second/2)/GameParams.CSize)); j++ {
			for coin, _ := range CoinMap[Pair[int,int]{int(i),int(j)}] {
				coins = append(coins, NearbyCoin{coin.Second.First, coin.Second.Second, coin.Second.Third})
			}
		}
	}

	return coins
}

func CheckExtraction(player ModPlayer) int {
	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting  player component: %w", err)
	}

	if playercomp.Coins > 0 && distance(playercomp.Loc, playercomp.Extract) <= ExtractionRadius {
		PlayerComp.Update(World, Players[player.Name], func(comp PlayerComponent) PlayerComponent{
			comp.Coins = 0// extraction point offloading

			return comp
		})

		return playercomp.Coins
	} else {
		return 0
	}
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
	PlayerMap[GetCell(player.Loc)][newPlayer] = pewp

	return nil
}

func RecentAttacks() []AttackTriple {
	return Attacks
}
