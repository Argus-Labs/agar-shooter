package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"

	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

func AddCoin(coin Triple[float64, float64, int]) (int, error) {
	coinID, err := World.Create(CoinComp)
	CoinComp.Set(World, coinID, CoinComponent{Pair[float64, float64]{coin.First, coin.Second}, coin.Third})

	if err != nil {
		return -1, fmt.Errorf("Coin creation failed: %w", err)
	}

	mutex.Lock()
	CoinMap[GetCell(coin)][Pair[storage.EntityID, Triple[float64, float64, int]]{coinID, coin}] = pewp
	mutex.Unlock()
	totalCoins++

	return coin.Third, nil
}

func RemoveCoin(coinID Pair[storage.EntityID, Triple[float64, float64, int]]) (int, error) {
	coin, err := CoinComp.Get(World, coinID.First)

	if err != nil {
		return -1, fmt.Errorf("Cardinal: could not get coin entity: %w", err)
	}

	mutex.Lock()
	delete(CoinMap[Pair[int,int]{int(math.Floor(coinID.Second.First/GameParams.CSize)),int(math.Floor(coinID.Second.Second/GameParams.CSize))}], coinID)
	mutex.Unlock()

	if err := World.Remove(coinID.First); err != nil {
		return -1, err
	}

	totalCoins--
	
	return coin.Val, nil
}

func PushPlayer(player PlayerComponent) error {
	if _, contains := Players[player.Name]; contains {// player already exists; don't do anything
		fmt.Println("Player already exists; not pushing again")
		return nil
	}

	playerID, err := World.Create(PlayerComp)// creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], player)

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting location with callback function: %w", err)
	}

	// adds player to kdtree
	PlayerTree.Insert(&P{vector.V{playercomp.Loc.First,playercomp.Loc.Second}, playercomp.Name})

	return nil
}

func PopPlayer(player PlayerComponent) error {
	delete(Players, player.Name)
	
	// removes player to kdtree; should only remove a single node
	point := &P{vector.V{player.Loc.First, player.Loc.Second}, player.Name}
	PlayerTree.Remove(point.P(), point.Equal)
	
	return nil
}

func HandlePlayerPush(player AddPlayer) error {
	playerComp := PlayerComponent{player.Name, 100, player.Coins, DefaultWeapon, Pair[float64,float64]{25 + (rand.Float64()-0.5)*10,25 + (rand.Float64()-0.5)*10}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, true, -1}
	//PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// default player
	return PushPlayer(playerComp)
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
		if _, err := AddCoin(coin); err != nil {
			return err
		}
	}

	return PopPlayer(playercomp)
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
		}
	}

	for _, playername := range GameParams.Players {
		playercomp := PlayerComponent{playername, 100, 0, DefaultWeapon, Pair[float64,float64]{25 + (rand.Float64()-0.5)*10,25 + (rand.Float64()-0.5)*10}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, true, -1}// initializes player entities through their component
		//PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// initializes player entitities through their component

		if err := PushPlayer(playercomp); err != nil {
			return err
		}
	}

	return nil
}

func SpawnCoins() error {// spawn coins randomly over the board until the coin cap has been met
	coinsToAdd := math.Min(float64(maxCoins() - totalCoins), float64(maxCoinsPerTick))

	for coinsToAdd > 0 {// generate coins if we haven't reached the max density
		newCoin := Triple[float64,float64,int]{coinRadius + rand.Float64()*(GameParams.Dims.First-2*coinRadius), coinRadius + rand.Float64()*(GameParams.Dims.Second-2*coinRadius), 1}// random location over range where coins can actually be generated
		keep := true
		coinRound := GetCell(newCoin)
		if len(CoinMap[coinRound]) >= maxCoinsInCell() { continue }

		for i := math.Max(0, float64(coinRound.First-1)); i <= math.Min(float64(Width), float64(coinRound.First+1)); i++ {
			for j := math.Max(0, float64(coinRound.Second-1)); i <= math.Min(float64(Height), float64(coinRound.Second+1)); i++ {
				mutex.RLock()
				for coin,_ := range CoinMap[Pair[int,int]{int(i), int(j)}] {
					keep = keep && (distance(coin.Second, newCoin) > 2*coinRadius)
				}
				mutex.RUnlock()

				knn := kd.KNN[*P](PlayerTree, vector.V{newCoin.First, newCoin.Second}, 1, func(q *P) bool {
					return true
				})

				if len(knn) > 0 {
					nearestPlayerComp, err := PlayerComp.Get(World, Players[knn[0].Name])

					if err != nil {
						return fmt.Errorf("Cardinal: player obtain: %w", err)
					}

					keep = keep && (distance(nearestPlayerComp.Loc, newCoin) > PlayerRadius+1+coinRadius)
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


func RecentAttacks() []AttackTriple {
	if rand.Float64() > 0.95 {
		PlayerTree.Balance()
	}
	return Attacks
}
