package systems

import (
	"fmt"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/types"
	"math"
	"math/rand"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

func bound(x float64, y float64) types.Pair[float64, float64] {
	return types.Pair[float64, float64]{
		First:  math.Min(float64(game.GameParams.Dims.First), math.Max(0, x)),
		Second: math.Min(float64(game.GameParams.Dims.Second), math.Max(0, y)),
	}
}

// returns distance between two coins
func distance(loc1, loc2 types.Mult) float64 {
	return math.Sqrt(math.Pow(loc1.GetFirst()-loc2.GetFirst(), 2) + math.Pow(loc1.GetSecond()-loc2.GetSecond(), 2))
}

// change speed function
func move(tmpPlayer components.PlayerComponent) types.Pair[float64, float64] {
	dir := tmpPlayer.Dir
	coins := tmpPlayer.Coins
	playerSpeed := float64(game.WorldConstants.PlayerSpeed)

	return bound(
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

	game.Mutex.Lock()
	game.CoinMap[types.GetCell(coin)][types.Pair[storage.EntityID, types.Triple[float64, float64, int]]{coinID, coin}] = types.Pewp
	game.Mutex.Unlock()
	game.TotalCoins++

	return coin.Third, nil
}

// attack a player
func attack(world *ecs.World, id storage.EntityID, weapon types.Weapon, left bool, attacker, defender string) error {
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
		loc = bound(loc.First+3*math.Cos(randfloat), loc.Second+3*math.Sin(randfloat))

		if _, err := AddCoin(world, types.Triple[float64, float64, int]{loc.First, loc.Second, 1}); err != nil {
			return err
		}

		game.Attacks = append(game.Attacks, types.AttackTriple{attacker, defender, -1})
	} else { // adds attack to display queue if it was executed
		game.Attacks = append(game.Attacks, types.AttackTriple{attacker, defender, worldConstants.Weapons[weapon].Attack})
	}

	// removes player from map if they die
	if kill {
		if err := HandlePlayerPop(types.ModPlayer{name}); err != nil {
			return err
		}
	}

	return nil
}

// moves player based on the coin-speed
func ProcessMovesSystem(World *ecs.World, q *ecs.TransactionQueue) error {
	attackQueue := make([]types.Triple[storage.EntityID, types.Weapon, types.Triple[bool, string, string]], 0)
	game.Attacks = make([]types.AttackTriple, 0)

	for playerName, id := range game.Players {
		tmpPlayer, err := components.Player.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		// attacking players; each player attacks the closest player TODO: change targetting system later

		var (
			minID             storage.EntityID
			minDistance       float64
			closestPlayerName string
			left              bool
		)

		assigned := false

		for _, closestPlayerID := range game.Players {
			if closestPlayerID != id {
				closestPlayer, err := components.Player.Get(World, closestPlayerID)
				if err != nil {
					return err
				}

				dist := distance(closestPlayer.Loc, prevLoc)

				if !assigned || minDistance > dist {
					minID = closestPlayerID
					minDistance = dist
					closestPlayerName = closestPlayer.Name
					assigned = true
					left = tmpPlayer.Loc.First <= closestPlayer.Loc.First
				}
			}
		}

		if assigned && minDistance <= game.Weapons[tmpPlayer.Weapon].Range {
			attackQueue = append(attackQueue, types.Triple[storage.EntityID, Weapon, types.Triple[bool, string, string]]{minID, tmpPlayer.Weapon, types.Triple[bool, string, string]{left, playerName, closestPlayerName}})
		}

		// moving players

		loc := move(tmpPlayer)

		delete(PlayerMap[GetCell(prevLoc)], types.Pair[storage.EntityID, types.Pair[float64, float64]]{id, prevLoc})
		PlayerMap[GetCell(loc)][types.Pair[storage.EntityID, types.Pair[float64, float64]]{id, loc}] = pewp

		hitCoins := make([]Pair[storage.EntityID, types.Triple[float64, float64, int]], 0)

		for i := int(math.Floor(prevLoc.First / GameParams.CSize)); i <= int(math.Floor(loc.First/GameParams.CSize)); i++ {
			for j := int(math.Floor(prevLoc.Second / GameParams.CSize)); j <= int(math.Floor(loc.Second/GameParams.CSize)); j++ {
				for coin, _ := range CoinMap[types.Pair[int, int]{i, j}] {
					if CoinProjDist(prevLoc, loc, coin.Second) <= PlayerRadius {
						hitCoins = append(hitCoins, coin)
					}
				}
			}
		}

		extraCoins := 0

		for _, entityID := range hitCoins {
			if coinVal, err := RemoveCoin(entityID); err != nil {
				return err
			} else {
				extraCoins += coinVal
			}

		}

		// modifies player location
		PlayerComp.Update(World, Players[playerName], func(comp PlayerComponent) PlayerComponent {
			comp.Loc = loc
			comp.Coins += extraCoins

			return comp
		})
	}

	for _, triple := range attackQueue {
		if err := attack(triple.First, triple.Second, triple.Third.First, triple.Third.Second, triple.Third.Third); err != nil {
			return err
		}
	}

	return nil
}
