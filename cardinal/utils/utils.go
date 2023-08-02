package utils

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/read"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
	
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"
)

func InitializeGame(world *ecs.World, gameParams types.Game) error {
	//if World.stateIsLoaded {
	//	return fmt.Errorf("already loaded state")
	//}
	rand.Seed(time.Now().UnixNano())
	if gameParams.CSize == 0 {
		return errors.New("Cardinal: CellSize is zero")
	}
	game.GameParams = gameParams

	game.Width = int(math.Ceil(game.GameParams.Dims.First / game.GameParams.CSize))
	game.Height = int(math.Ceil(game.GameParams.Dims.Second / game.GameParams.CSize))

	for i := 0; i <= game.Width; i++ {
		for j := 0; j <= game.Height; j++ {
			game.CoinMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void)
			game.HealthMap[types.Pair[int, int]{i, j}] = make(map[types.Pair[storage.EntityID, types.Triple[float64, float64, int]]]types.Void)
		}
	}

	for i := 0; i < game.WorldConstants.InitRepeatSpawn; i++ {
		go SpawnCoins(world)
	}

	return nil
}

func SpawnCoins(world *ecs.World) error { // spawn coins randomly over the board until the coin cap has been met
	consts := game.WorldConstants
	coinsToAdd := math.Min(float64(game.MaxCoins()-game.TotalCoins), float64(consts.MaxCoinsPerTick))

	for coinsToAdd > 0 { // generate coins if we haven't reached the max density
		newCoin := types.Triple[float64, float64, int]{consts.CoinRadius + rand.Float64()*(game.GameParams.Dims.First-2*consts.CoinRadius), consts.CoinRadius + rand.Float64()*(game.GameParams.Dims.Second-2*consts.CoinRadius), 1} // random location over range where coins can actually be generated
		keep := true
		coinRound := GetCell(newCoin)
		if len(game.CoinMap[coinRound]) >= game.MaxCoinsInCell() {
			continue
		}

		for i := math.Max(0, float64(coinRound.First-1)); i <= math.Min(float64(game.Width), float64(coinRound.First+1)); i++ {
			for j := math.Max(0, float64(coinRound.Second-1)); j <= math.Min(float64(game.Height), float64(coinRound.Second+1)); j++ {
				game.CoinMutex.RLock()
				for coin, _ := range game.CoinMap[types.Pair[int, int]{int(i), int(j)}] {
					keep = keep && (Distance(coin.Second, newCoin) > 2*consts.CoinRadius)
				}
				game.CoinMutex.RUnlock()

				knn := kd.KNN[*types.P](game.PlayerTree, vector.V{newCoin.First, newCoin.Second}, 1, func(q *types.P) bool {
					return true
				})

				if len(knn) > 0 {
					if nearestPlayerComp, err := components.Player.Get(world, game.Players[knn[0].Name]); err != nil {
						return fmt.Errorf("Cardinal: player obtain: %w", err)
					} else {
						keep = keep && (Distance(nearestPlayerComp.Loc, newCoin) > consts.PlayerRadius + 1 + consts.CoinRadius)
					}
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

func AddPlayer(world *ecs.World, playerName string, playerCoins int) error {
	// Check whether the player exists
	if _, contains := game.Players[playerName]; contains {
		return fmt.Errorf("Cardinal: cannot add player with duplicate name")
	}

	game.PlayerCoins[playerName] = 0

	// Create the player
	playerID, err := world.Create(components.Player)

	if err != nil {
		return fmt.Errorf("Error adding player to world:", err)
	}

	game.Players[playerName] = playerID

	// Set the component to the correct values
	weaponID, err := world.Create(components.Weapon)
	components.Weapon.Set(world, weaponID, components.WeaponComponent{
		Loc: types.Pair[float64, float64]{-1, -1},
		Val: game.DefaultWeapon,
		Ammo: game.WorldConstants.Weapons[game.DefaultWeapon].MaxAmmo,
		LastAttack: 0,
	})

	components.Player.Set(world, playerID, components.PlayerComponent{
		Name: playerName,
		Health: 100,
		Coins: playerCoins,
		Weapon: weaponID,
		Loc: types.Pair[float64, float64]{0, 0},
		Dir: types.Pair[float64, float64]{0, 0},
		LastMove: types.Pair[float64, float64]{0, 0},
		IsRight: false,
		MoveNum: -1,
		Level: 0,
	})

	// Add player to local PlayerTree
	playerComp, err := components.Player.Get(world, playerID)
	game.PlayerTree.Insert(&types.P{vector.V{playerComp.Loc.First, playerComp.Loc.Second}, playerComp.Name})
	log.Debug().Msgf("Created player with name", playerComp.Name)

	return nil
}

func RemovePlayer(world *ecs.World, playerName string, playerList []read.PlayerPair) error {
	// Check that the player exists
	if _, contains := game.Players[playerName]; !contains {
		log.Error().Msg("player name does not exist")
		return fmt.Errorf("Cardinal: cannot remove player that does not exist")
	}

	// Get the player id and component
	player, err := read.GetPlayerByName(world, playerName)
	if err != nil {
		return err
	}

	// Remove the player from the World
	if err := world.Remove(player.ID); err != nil {
		return fmt.Errorf("RemovePlayerSystem: Error removing player", err)
	}

	// Put all the coins around the player
	coins := player.Component.Coins
	tot := int(math.Max(1, float64(coins/10 + (coins%10)/5 + coins%5)))
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

		peep := Bound(player.Component.Loc.First + rad*math.Cos(2*math.Pi*float64(start)/float64(tot)), player.Component.Loc.Second + rad*math.Sin(2*math.Pi*float64(start)/float64(tot)))
		newCoins = append(newCoins, types.Triple[float64, float64, int]{peep.First, peep.Second, addCoins})
		start++
	}

	for _, coin := range newCoins {
		if _, err := AddCoin(world, coin); err != nil {
			return err
		}
	}

	if _, err := AddHealth(world, types.Triple[float64, float64, int]{player.Component.Loc.First, player.Component.Loc.Second, player.Component.Health}); err != nil {
		return err
	}

	// Delete the player from the local PlayerTree
	delete(game.Players, player.Component.Name)

	point := &types.P{vector.V{player.Component.Loc.First, player.Component.Loc.Second}, player.Component.Name}
	game.PlayerTree.Remove(point.P(), point.Equal)

	return err
}

func Bound(x float64, y float64) types.Pair[float64, float64] {
	return types.Pair[float64, float64]{
		First:  math.Min(float64(game.GameParams.Dims.First), math.Max(0, x)),
		Second: math.Min(float64(game.GameParams.Dims.Second), math.Max(0, y)),
	}
}

// returns Distance between two coins
func Distance(loc1, loc2 types.Mult) float64 {
	return math.Sqrt(math.Pow(loc1.GetFirst()-loc2.GetFirst(), 2) + math.Pow(loc1.GetSecond()-loc2.GetSecond(), 2))
}

// contains speed function
func Move(tmpPlayer components.PlayerComponent) types.Pair[float64, float64] {
	dir := tmpPlayer.Dir
	coins := tmpPlayer.Coins
	playerSpeed := float64(game.WorldConstants.PlayerSpeed)

	return Bound(
		tmpPlayer.Loc.First+(playerSpeed*dir.First*math.Exp(-0.01*float64(coins))),
		tmpPlayer.Loc.Second+(playerSpeed*dir.Second*math.Exp(-0.01*float64(coins))),
	)
}

// closest distance the coin is from the player obtained by checking the orthogonal projection of the coin with the segment defiend by [start.end]
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

	if proj.First*vec.First+proj.Second*vec.Second < 0 || proj.First*proj.First+proj.Second*proj.Second > vec.First*vec.First+vec.Second*vec.Second {// outside span of [start, end]
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

	game.CoinMutex.Lock()
	game.CoinMap[GetCell(coin)][types.Pair[storage.EntityID, types.Triple[float64, float64, int]]{coinID, coin}] = types.Pewp
	game.CoinMutex.Unlock()
	game.TotalCoins++

	return coin.Third, nil
}

func RemoveCoin(world *ecs.World, coinID types.Pair[storage.EntityID, types.Triple[float64, float64, int]]) (int, error) {
	coin, err := components.Coin.Get(world, coinID.First)

	if err != nil {
		return -1, fmt.Errorf("Cardinal: could not get coin entity: %w", err)
	}

	game.CoinMutex.Lock()
	delete(game.CoinMap[types.Pair[int, int]{int(math.Floor(coinID.Second.First / game.GameParams.CSize)), int(math.Floor(coinID.Second.Second / game.GameParams.CSize))}], coinID)
	game.CoinMutex.Unlock()

	if err := world.Remove(coinID.First); err != nil {
		return -1, err
	}

	game.TotalCoins--

	return coin.Val, nil
}

func AddHealth(world *ecs.World, health types.Triple[float64, float64, int]) (int, error) {
	healthID, err := world.Create(components.Health)

	if err != nil {
		return -1, fmt.Errorf("Health creation failed: %w", err)
	}

	components.Health.Set(world, healthID, components.HealthComponent{types.Pair[float64, float64]{health.First, health.Second}, health.Third})

	game.HealthMutex.Lock()
	game.HealthMap[GetCell(health)][types.Pair[storage.EntityID, types.Triple[float64, float64, int]]{healthID, health}] = types.Pewp
	game.HealthMutex.Unlock()
	game.TotalHealth++

	return health.Third, nil
}

func RemoveHealth(world *ecs.World, healthID types.Pair[storage.EntityID, types.Triple[float64, float64, int]]) (int, error) {
	health, err := components.Health.Get(world, healthID.First)

	if err != nil {
		return -1, fmt.Errorf("Cardinal: could not get health entity: %w", err)
	}

	game.HealthMutex.Lock()
	delete(game.HealthMap[types.Pair[int, int]{int(math.Floor(healthID.Second.First / game.GameParams.CSize)), int(math.Floor(healthID.Second.Second / game.GameParams.CSize))}], healthID)
	game.HealthMutex.Unlock()

	if err := world.Remove(healthID.First); err != nil {
		return -1, fmt.Errorf("Cardinal: error removing health entity:", err)
	}

	game.TotalHealth--

	return health.Val, nil
}

func Attack(world *ecs.World, id, weapon storage.EntityID, left bool, attacker, defender string) error {
	wipun, err := components.Weapon.Get(world, weapon)

	if err != nil {
		return fmt.Errorf("Cardinal: error fetching weapon:", err)
	}

	if wipun.Ammo == 0 || wipun.Val == game.Dud || (wipun.LastAttack + game.WorldConstants.Weapons[wipun.Val].Reload) > time.Now().UnixNano() {
		return nil
	}

	kill := false
	coins := false
	var loc types.Pair[float64, float64]
	var name string
	worldConstants := game.WorldConstants

	if err := components.Player.Update(world, id, func(comp components.PlayerComponent) components.PlayerComponent { // modifies player location
		if left == comp.IsRight && comp.Coins > 0 {
			comp.Coins--
			coins = true
		} else {
			if attacker_, err := components.Player.Get(world, game.Players[attacker]); err == nil {
				comp.Health -= int(math.Floor(float64(worldConstants.Weapons[wipun.Val].Attack) * (1 + game.LevelAttack[attacker_.Level])))
			}
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
		loc = Bound(loc.First + 3*math.Cos(randfloat), loc.Second + 3*math.Sin(randfloat))

		if _, err := AddCoin(world, types.Triple[float64, float64, int]{First: loc.First, Second: loc.Second, Third: 1}); err != nil {
			return err
		}

		game.Attacks = append(game.Attacks, types.AttackTriple{AttackerID: attacker, DefenderID: defender, Damage: -1})
	} else { // adds attack to display queue if it was executed
		game.Attacks = append(game.Attacks, types.AttackTriple{AttackerID: attacker, DefenderID: defender, Damage: worldConstants.Weapons[wipun.Val].Attack})
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

func GetCell(loc types.Mult) types.Pair[int, int] {
	cellSize := game.GameParams.CSize
	return types.Pair[int, int]{int(math.Floor(loc.GetFirst() / cellSize)), int(math.Floor(loc.GetSecond() / cellSize))}
}
