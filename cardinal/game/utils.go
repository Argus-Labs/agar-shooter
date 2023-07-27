package game

import (
	"errors"
	"fmt"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/read"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
	"math"
	"math/rand"
	"time"
)

func InitializeGame(world *ecs.World, gameParams types.Game) error {
	//if World.stateIsLoaded {
	//	return fmt.Errorf("already loaded state")
	//}
	rand.Seed(time.Now().UnixNano())
	if gameParams.CSize == 0 {
		return errors.New("Cardinal: CellSize is zero")
	}
	GameParams = gameParams

	Width = int(math.Ceil(GameParams.Dims.First / GameParams.CSize))
	Height = int(math.Ceil(GameParams.Dims.Second / GameParams.CSize))

	for i := 0; i < WorldConstants.InitRepeatSpawn; i++ {
		go SpawnCoins(world)
	}

	return nil
}

func SpawnCoins(world *ecs.World) error { // spawn coins randomly over the board until the coin cap has been met
	consts := WorldConstants
	coinsToAdd := math.Min(float64(MaxCoins()-TotalCoins), float64(consts.MaxCoinsPerTick))

	for coinsToAdd > 0 { // generate coins if we haven't reached the max density
		newCoin := types.Triple[float64, float64, int]{consts.CoinRadius + rand.Float64()*(GameParams.Dims.First-2*consts.CoinRadius), consts.CoinRadius + rand.Float64()*(GameParams.Dims.Second-2*consts.CoinRadius), 1} // random location over range where coins can actually be generated
		keep := true
		coinRound := types.GetCell(newCoin)
		if len(CoinMap[coinRound]) >= MaxCoinsInCell() {
			continue
		}

		for i := math.Max(0, float64(coinRound.First-1)); i <= math.Min(float64(Width), float64(coinRound.First+1)); i++ {
			for j := math.Max(0, float64(coinRound.Second-1)); i <= math.Min(float64(Height), float64(coinRound.Second+1)); i++ {
				Mutex.RLock()
				for coin, _ := range CoinMap[types.Pair[int, int]{int(i), int(j)}] {
					keep = keep && (Distance(coin.Second, newCoin) > 2*consts.CoinRadius)
				}
				Mutex.RUnlock()

				for player, _ := range PlayerMap[types.Pair[int, int]{int(i), int(j)}] {
					keep = keep && (Distance(player.Second, newCoin) > consts.PlayerRadius+1+consts.CoinRadius)
				}
			}
		}
		if keep {
			if _, err := AddCoin(world, newCoin); err != nil {
				return err
			}

			coinsToAdd--
		}
	}

	// create mutex to prevent concurrent ticks from causing problems; iterating through map above takes too much time to do, so when the second tick is called and iteration occurs, the first tick is still trying to add elements to the map
	// also limit the number of coins in each cell of the coinmap and the size of the map so we don't have iteration problems
	// maybe make this a system so it can be run async

	return nil
}

func RemovePlayer(world *ecs.World, playerName string, playerList []read.PlayerPair) error {
	// Check that the player exists
	var playerFound bool = false
	for _, player := range playerList {
		if player.Component.Name == playerName {
			playerFound = true
		}
	}
	if playerFound == false {
		log.Error().Msg("player name already exists")
		return errors.New("RemovePlayerSystem: Player does not exist")
	}

	// Get the player id and component
	player, err := read.GetPlayerByName(world, playerName)
	if err != nil {
		return err
	}

	// Remove the player from the World
	if err := world.Remove(player.ID); err != nil {
		log.Error().Msg("RemovePlayerSystem: Error removing player from world")
	}

	// Put all the coins around the player
	coins := player.Component.Coins
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

		peep := Bound(player.Component.Loc.First+rad*math.Cos(2*math.Pi*float64(start)/float64(tot)), player.Component.Loc.Second+rad*math.Sin(2*math.Pi*float64(start)/float64(tot)))
		newCoins = append(newCoins, types.Triple[float64, float64, int]{peep.First, peep.Second, addCoins})
		start++
	}

	for _, coin := range newCoins {
		if _, err := AddCoin(world, coin); err != nil {
			return err
		}
	}

	// Delete the player from the local PlayerMap
	oldPlayer := types.Pair[storage.EntityID, types.Pair[float64, float64]]{player.ID, player.Component.Loc}
	delete(PlayerMap[types.GetCell(player.Component.Loc)], oldPlayer)

	delete(Players, player.Component.Name)

	return err
}

func Bound(x float64, y float64) types.Pair[float64, float64] {
	return types.Pair[float64, float64]{
		First:  math.Min(float64(GameParams.Dims.First), math.Max(0, x)),
		Second: math.Min(float64(GameParams.Dims.Second), math.Max(0, y)),
	}
}

// returns Distance between two coins
func Distance(loc1, loc2 types.Mult) float64 {
	return math.Sqrt(math.Pow(loc1.GetFirst()-loc2.GetFirst(), 2) + math.Pow(loc1.GetSecond()-loc2.GetSecond(), 2))
}

// change speed function
func Move(tmpPlayer components.PlayerComponent) types.Pair[float64, float64] {
	dir := tmpPlayer.Dir
	coins := tmpPlayer.Coins
	playerSpeed := float64(WorldConstants.PlayerSpeed)

	return Bound(
		tmpPlayer.Loc.First+(playerSpeed*dir.First*math.Exp(-0.01*float64(coins))),
		tmpPlayer.Loc.Second+(playerSpeed*dir.Second*math.Exp(-0.01*float64(coins))),
	)
}

func CoinProjDist(start, end types.Pair[float64, float64], coin types.Triple[float64, float64, int]) float64 {
	vec := types.Pair[float64, float64]{
		First:  end.First - start.First,
		Second: end.Second - start.Second,
	}
	coin = types.Triple[float64, float64, int]{
		First:  coin.First - start.First,
		Second: coin.Second - start.Second,
		Third:  0,
	}
	coeff := (vec.First*coin.First + vec.Second*coin.Second) / (vec.First*vec.First + vec.Second*vec.Second)
	proj := types.Pair[float64, float64]{
		First:  coeff*vec.First + start.First,
		Second: coeff*vec.Second + start.Second,
	}
	ortho := types.Pair[float64, float64]{
		First:  coin.First - proj.First,
		Second: coin.Second - proj.Second,
	}

	if proj.First*vec.First+proj.Second*vec.Second < 0 || proj.First*proj.First+proj.Second*proj.Second > vec.First*vec.First+vec.Second*vec.Second {
		return math.Sqrt(math.Min(coin.First*coin.First+coin.Second*coin.Second, (coin.First-vec.First)*(coin.First-vec.First)+(coin.Second-vec.Second)*(coin.Second-vec.Second)))
	} else {
		return math.Sqrt(ortho.First*ortho.First + ortho.Second*ortho.Second)
	}
}

func AddCoin(world *ecs.World, coin types.Triple[float64, float64, int]) (int, error) {
	coinID, err := world.Create(components.Coin)
	components.Coin.Set(world, coinID, components.CoinComponent{types.Pair[float64, float64]{coin.First, coin.Second}, coin.Third})

	if err != nil {
		return -1, fmt.Errorf("Coin creation failed: %w", err)
	}

	Mutex.Lock()
	CoinMap[types.GetCell(coin)][types.Pair[storage.EntityID, types.Triple[float64, float64, int]]{coinID, coin}] = types.Pewp
	Mutex.Unlock()
	TotalCoins++

	return coin.Third, nil
}

func RemoveCoin(world *ecs.World, coinID types.Pair[storage.EntityID, types.Triple[float64, float64, int]]) (int, error) {
	coin, err := components.Coin.Get(world, coinID.First)

	if err != nil {
		return -1, fmt.Errorf("Cardinal: could not get coin entity: %w", err)
	}

	Mutex.Lock()
	delete(CoinMap[types.Pair[int, int]{int(math.Floor(coinID.Second.First / GameParams.CSize)), int(math.Floor(coinID.Second.Second / GameParams.CSize))}], coinID)
	Mutex.Unlock()

	if err := world.Remove(coinID.First); err != nil {
		return -1, err
	}

	TotalCoins--

	return coin.Val, nil
}

func Attack(world *ecs.World, id storage.EntityID, weapon types.Weapon, left bool, attacker, defender string) error {
	kill := false
	coins := false
	var loc types.Pair[float64, float64]
	var name string
	worldConstants := WorldConstants

	if err := components.Player.Update(world, id, func(comp components.PlayerComponent) components.PlayerComponent { // modifies player location
		if left == comp.IsRight && comp.Coins > 0 {
			comp.Coins--
			coins = true
		} else {
			comp.Health -= worldConstants.Weapons[weapon].Attack
		}
		kill = comp.Health <= 0
		name = comp.Name
		loc = comp.Loc

		return comp
	}); err != nil {
		return nil
	}

	if coins {
		randfloat := rand.Float64() * 2 * math.Pi
		loc = Bound(loc.First+3*math.Cos(randfloat), loc.Second+3*math.Sin(randfloat))

		if _, err := AddCoin(world, types.Triple[float64, float64, int]{First: loc.First, Second: loc.Second, Third: 1}); err != nil {
			return err
		}

		Attacks = append(Attacks, types.AttackTriple{AttackerID: attacker, DefenderID: defender, Damage: -1})
	} else { // adds attack to display queue if it was executed
		Attacks = append(Attacks, types.AttackTriple{AttackerID: attacker, DefenderID: defender, Damage: worldConstants.Weapons[weapon].Attack})
	}

	// removes player from map if they die
	if kill {
		playerList := read.ReadPlayers(world)
		if err := RemovePlayer(world, name, playerList); err != nil {
			return err
		}
	}

	return nil
}
