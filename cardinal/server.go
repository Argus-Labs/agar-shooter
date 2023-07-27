package main

import (
	"fmt"
	"github.com/argus-labs/new-game/components"
	"math"
	"math/rand"
	"time"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

func RemoveCoin(coinID types.Pair[storage.EntityID, types.Triple[float64, float64, int]]) (int, error) {
	coin, err := CoinComp.Get(World, coinID.First)

	if err != nil {
		return -1, fmt.Errorf("Cardinal: could not get coin entity: %w", err)
	}

	mutex.Lock()
	delete(CoinMap[types.Pair[int, int]{int(math.Floor(coinID.Second.First / GameParams.CSize)), int(math.Floor(coinID.Second.Second / GameParams.CSize))}], coinID)
	mutex.Unlock()

	if err := World.Remove(coinID.First); err != nil {
		return -1, err
	}

	totalCoins--

	return coin.Val, nil
}

func TickTock() error { // testing function used to make the game tick
	err := World.Tick()
	return err
}

func CreateGame(game Game) error {
	//if World.stateIsLoaded {
	//	return fmt.Errorf("already loaded state")
	//}
	rand.Seed(time.Now().UnixNano())
	if game.CSize == 0 {
		return fmt.Errorf("Cardinal: cellsize is zero")
	}
	GameParams = game
	World.RegisterComponents(PlayerComp, CoinComp, HealthComp, WeaponComp)
	World.AddSystem(processMoves)
	World.AddSystem(makeMoves)

	World.LoadGameState()
	MoveTx.SetID(0)
	playerIDs, err := World.CreateMany(len(GameParams.Players), PlayerComp) // creates player entities

	Players = make(map[string]storage.EntityID)
	for i, playername := range GameParams.Players { // associates storage.EntityIDs with each player
		Players[playername] = playerIDs[i]
	}

	if err != nil {
		return fmt.Errorf("Error initializing game objects: %w", err)
	}

	Width = int(math.Ceil(GameParams.Dims.First / GameParams.CSize))
	Height = int(math.Ceil(GameParams.Dims.Second / GameParams.CSize))

	// initializes player and item maps
	for i := 0; i <= Width; i++ {
		for j := 0; j <= Height; j++ {
			CoinMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]void)
			HealthMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]void)
			WeaponMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]void)
			PlayerMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]void)
		}
	}

	for _, playername := range GameParams.Players {
		PlayerComp.Set(World, Players[playername], components.PlayerComponent{playername, 100, 0, DefaultWeapon, types.Pair[float64, float64]{25 + (rand.Float64()-0.5)*10, 25 + (rand.Float64()-0.5)*10}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{rand.Float64() * GameParams.Dims.First, rand.Float64() * GameParams.Dims.Second}, true, -1}) // initializes player entities through their component
		//PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Dud, types.Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, types.Pair[float64,float64]{0,0}, types.Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// initializes player entitities through their component

		playercomp, err := PlayerComp.Get(World, Players[playername])

		if err != nil {
			fmt.Errorf("Cardinal: Error getting location with callback function: %w", err)
		}

		newPlayer := types.Pair[storage.EntityID, types.Pair[float64, float64]]{Players[playername], playercomp.Loc}
		PlayerMap[GetCell(playercomp.Loc)][newPlayer] = pewp
	}

	return nil
}

func SpawnCoins() error { // spawn coins randomly over the board until the coin cap has been met
	coinsToAdd := math.Min(float64(maxCoins()-totalCoins), float64(maxCoinsPerTick))

	for coinsToAdd > 0 { // generate coins if we haven't reached the max density
		newCoin := types.Triple[float64, float64, int]{coinRadius + rand.Float64()*(GameParams.Dims.First-2*coinRadius), coinRadius + rand.Float64()*(GameParams.Dims.Second-2*coinRadius), 1} // random location over range where coins can actually be generated
		keep := true
		coinRound := GetCell(newCoin)
		if len(CoinMap[coinRound]) >= maxCoinsInCell() {
			continue
		}

		for i := math.Max(0, float64(coinRound.First-1)); i <= math.Min(float64(Width), float64(coinRound.First+1)); i++ {
			for j := math.Max(0, float64(coinRound.Second-1)); i <= math.Min(float64(Height), float64(coinRound.Second+1)); i++ {
				mutex.RLock()
				for coin, _ := range CoinMap[types.Pair[int, int]{int(i), int(j)}] {
					keep = keep && (distance(coin.Second, newCoin) > 2*coinRadius)
				}
				mutex.RUnlock()

				for player, _ := range PlayerMap[types.Pair[int, int]{int(i), int(j)}] {
					keep = keep && (distance(player.Second, newCoin) > PlayerRadius+1+coinRadius)
				}
			}
		}
		if keep {
			if _, err := AddCoin(newCoin); err != nil {
				return err
			}

			coinsToAdd--
		}
	}

	//create mutex to prevent concurrent ticks from causing problems; iterating through map above takes too much time to do, so when the second tick is called and iteration occurs, the first tick is still trying to add elements to the map
	// also limit the number of coins in each cell of the coinmap and the size of the map so we don't have iteration problems
	// maybe make this a system so it can be run async

	return nil
}

func CheckExtraction(player ModPlayer) int {
	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting  player component: %w", err)
	}

	if playercomp.Coins > 0 && distance(playercomp.Loc, playercomp.Extract) <= ExtractionRadius {
		PlayerComp.Update(World, Players[player.Name], func(comp components.PlayerComponent) components.PlayerComponent {
			comp.Coins = 0 // extraction point offloading

			return comp
		})

		return playercomp.Coins
	} else {
		return 0
	}
}

func AddTestPlayer(player components.PlayerComponent) error {
	if _, contains := Players[player.Name]; contains { // player already exists; don't do anything
		return nil
	}

	playerID, err := World.Create(PlayerComp) // creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], player) // default player

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		return fmt.Errorf("Error getting location with callback function: %w", err)
	}

	newPlayer := types.Pair[storage.EntityID, types.Pair[float64, float64]]{Players[player.Name], playercomp.Loc}
	PlayerMap[GetCell(player.Loc)][newPlayer] = pewp

	return nil
}

func RecentAttacks() []AttackTriple {
	return Attacks
}
