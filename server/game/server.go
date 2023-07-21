package game

import (
	"fmt"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/component"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/systems"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/types"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"
	"math"
	"math/rand"
	"time"

	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

func AddCoin(coin types.Triple[float64, float64, int]) (int, error) {
	coinID, err := World.Create(CoinComp)
	CoinComp.Set(World, coinID, component.CoinComponent{types.Pair[float64, float64]{coin.First, coin.Second}, coin.Third})

	if err != nil {
		return -1, fmt.Errorf("Coin creation failed: %w", err)
	}

	mutex.Lock()
	CoinMap[types.GetCell(coin)][types.Pair[storage.EntityID, types.Triple[float64, float64, int]]{coinID, coin}] = types.Pewp
	mutex.Unlock()
	totalCoins++

	return coin.Third, nil
}

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

func PushPlayer(player component.PlayerComponent) error {
	if _, contains := Players[player.Name]; contains { // player already exists; don't do anything
		fmt.Println("Player already exists; not pushing again")
		return nil
	}

	//TODO remove after demo
	PlayerMaxCoins[player.Name] = 0

	playerID, err := World.Create(PlayerComp) // creates new player
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
	PlayerTree.Insert(&types.P{vector.V{playercomp.Loc.First, playercomp.Loc.Second}, playercomp.Name})

	return nil
}

func PopPlayer(player component.PlayerComponent) error {
	delete(Players, player.Name)

	// removes player to kdtree; should only remove a single node
	point := &types.P{vector.V{player.Loc.First, player.Loc.Second}, player.Name}
	PlayerTree.Remove(point.P(), point.Equal)

	return nil
}

func HandlePlayerPushInternal(player types.AddPlayer) error {
	playerComp := component.PlayerComponent{player.Name, 100, player.Coins, DefaultWeapon, types.Pair[float64, float64]{25 + (rand.Float64()-0.5)*10, 25 + (rand.Float64()-0.5)*10}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{rand.Float64() * GameParams.Dims.First, rand.Float64() * GameParams.Dims.Second}, true, -1}
	//PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// default player
	return PushPlayer(playerComp)
}

func HandlePlayerPopInternal(player types.ModPlayer) error {
	playercomp, err := PlayerComp.Get(World, Players[player.Name])
	if err != nil {
		return err
	}

	if err = World.Remove(Players[player.Name]); err != nil {
		fmt.Errorf("error removing player: %w", err)
	}

	// put all coins around the player
	coins := playercomp.Coins
	tot := int(math.Max(1, float64(coins/10+(coins%10)/5+coins%5)))
	start := 0
	rad := float64(tot) / (2 * math.Pi)
	newCoins := make([]types.Triple[float64, float64, int], 0)

	for coins > 0 { // decomposes into 10s, 5s, 1s
		addCoins := 0
		switch {
		case coins >= 10:
			{
				addCoins = 10
				coins -= 10
				break
			}
		case coins >= 5:
			{
				addCoins = 5
				coins -= 5
				break
			}
		default:
			{
				addCoins = 1
				coins--
			}
		}

		peep := systems.Bound(playercomp.Loc.First+rad*math.Cos(2*math.Pi*float64(start)/float64(tot)), playercomp.Loc.Second+rad*math.Sin(2*math.Pi*float64(start)/float64(tot)))
		newCoins = append(newCoins, types.Triple[float64, float64, int]{peep.First, peep.Second, addCoins})
		start++
	}

	for _, coin := range newCoins {
		if _, err := AddCoin(coin); err != nil {
			return err
		}
	}

	return PopPlayer(playercomp)
}

func TickTock() error { // testing function used to make the game tick
	err := World.Tick()
	return err
}

func GetPlayerStateInternal(player types.ModPlayer) (component.PlayerComponent, error) { // testing function used in place of broadcast to get state of players
	if _, contains := Players[player.Name]; contains == false {
		return component.PlayerComponent{}, fmt.Errorf("Player does not exist")
	}

	comp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		return component.PlayerComponent{}, fmt.Errorf("Player fetch error: %w", err)
	}

	return comp, nil
}

func GetPlayerStatusInternal() []types.Pair[string, types.Pair[float64, float64]] { // sends all player information to each player
	locs := make([]types.Pair[string, types.Pair[float64, float64]], 0)
	for key, id := range Players {
		comp, _ := PlayerComp.Get(World, id)
		locs = append(locs, types.Pair[string, types.Pair[float64, float64]]{key, comp.Loc})
	}

	return locs
}

func HandleMakeMoveInternal(move types.Move) {
	MoveTx.AddToQueue(World, move) // adds "move" transaction to World transaction queue
}

func CreateGameInternal(_game types.Game) error {
	//if World.stateIsLoaded {
	//	return fmt.Errorf("already loaded state")
	//}
	rand.Seed(time.Now().UnixNano())
	if _game.CSize == 0 {
		return fmt.Errorf("Cardinal: cellsize is zero")
	}
	GameParams = _game
	World.RegisterComponents(PlayerComp, CoinComp, HealthComp, WeaponComp)
	World.AddSystem(systems.ProcessMoves)
	World.AddSystem(systems.MakeMoves)

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
			CoinMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void)
			HealthMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.Void)
			WeaponMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Pair[float64, float64]]]types.Void)
		}
	}

	for _, playername := range GameParams.Players {
		playercomp := component.PlayerComponent{playername, 100, 0, DefaultWeapon, types.Pair[float64, float64]{25 + (rand.Float64()-0.5)*10, 25 + (rand.Float64()-0.5)*10}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{rand.Float64() * GameParams.Dims.First, rand.Float64() * GameParams.Dims.Second}, true, -1} // initializes player entities through their component
		//PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Dud, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{rand.Float64()*GameParams.Dims.First, rand.Float64()*GameParams.Dims.Second}, -1})// initializes player entitities through their component

		if err := PushPlayer(playercomp); err != nil {
			return err
		}
	}

	return nil
}

func SpawnCoins() error { // spawn coins randomly over the board until the coin cap has been met
	coinsToAdd := math.Min(float64(maxCoins()-totalCoins), float64(maxCoinsPerTick))

	for coinsToAdd > 0 { // generate coins if we haven't reached the max density
		newCoin := types.Triple[float64, float64, int]{coinRadius + rand.Float64()*(GameParams.Dims.First-2*coinRadius), coinRadius + rand.Float64()*(GameParams.Dims.Second-2*coinRadius), 1} // random location over range where coins can actually be generated
		keep := true
		coinRound := types.GetCell(newCoin)
		if len(CoinMap[coinRound]) >= maxCoinsInCell() {
			continue
		}

		for i := math.Max(0, float64(coinRound.First-1)); i <= math.Min(float64(Width), float64(coinRound.First+1)); i++ {
			for j := math.Max(0, float64(coinRound.Second-1)); i <= math.Min(float64(Height), float64(coinRound.Second+1)); i++ {
				mutex.RLock()
				for coin, _ := range CoinMap[types.Pair[int, int]{int(i), int(j)}] {
					keep = keep && (systems.Distance(coin.Second, newCoin) > 2*coinRadius)
				}
				mutex.RUnlock()

				knn := kd.KNN[*types.P](PlayerTree, vector.V{newCoin.First, newCoin.Second}, 1, func(q *types.P) bool {
					return true
				})

				if len(knn) > 0 {
					nearestPlayerComp, err := PlayerComp.Get(World, Players[knn[0].Name])

					if err != nil {
						return fmt.Errorf("Cardinal: player obtain: %w", err)
					}

					keep = keep && (systems.Distance(nearestPlayerComp.Loc, newCoin) > PlayerRadius+1+coinRadius)
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

func NearbyCoins(player types.ModPlayer) []types.NearbyCoin {
	coins := make([]types.NearbyCoin, 0)

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting player component: %w", err)
	}

	for i := math.Max(0, math.Floor((playercomp.Loc.First-ClientView.First/2)/GameParams.CSize)); i <= math.Min(float64(Width), math.Ceil((playercomp.Loc.First+ClientView.First/2)/GameParams.CSize)); i++ {
		for j := math.Max(0, math.Floor((playercomp.Loc.Second-ClientView.Second/2)/GameParams.CSize)); j <= math.Min(float64(Height), math.Ceil((playercomp.Loc.Second+ClientView.Second/2)/GameParams.CSize)); j++ {
			for coin, _ := range CoinMap[types.Pair[int, int]{int(i), int(j)}] {
				coins = append(coins, types.NearbyCoin{coin.Second.First, coin.Second.Second, coin.Second.Third})
			}
		}
	}

	return coins
}

func GetExtractionPointInternal(player types.ModPlayer) types.Pair[float64, float64] {
	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting player component: %w", err)
	}

	return playercomp.Extract
}

func CheckExtractionInternal(player types.ModPlayer) int {
	return PlayerMaxCoins[player.Name] //TODO remove after demo
	/*
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
	*/
}

func RecentAttacksInternal() []types.AttackTriple {
	if rand.Float64() > 0.95 {
		PlayerTree.Balance()
	}
	return Attacks
}
