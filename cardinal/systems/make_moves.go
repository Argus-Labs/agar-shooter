package systems

import (
	"math"
	"fmt"

	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"

	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

// moves player based on the coin-speed
func ProcessMovesSystem(world *ecs.World, q *ecs.TransactionQueue) error {
	attackQueue := make([]types.Triple[storage.EntityID, storage.EntityID, types.Triple[bool, string, string]], 0)
	game.Attacks = make([]types.AttackTriple, 0)
	maxDepth := 0

	for personaTag, id := range game.Players {
		tmpPlayer, err := components.Player.Get(world, id)

		if err != nil {
			return err
		}

		weapon, err := components.Weapon.Get(world, tmpPlayer.Weapon)

		if err != nil {
			return err
		}

		attackRange := game.WorldConstants.Weapons[weapon.Val].Range
		prevLoc := tmpPlayer.Loc

		// attacking players; each player attacks the closest player
		depth := 0
		knn := kd.KNN[*types.P](game.PlayerTree, vector.V{prevLoc.First, prevLoc.Second}, 2, func(q *types.P) bool {
			depth++
			return true
		})

		if maxDepth < depth { maxDepth = depth }

		if len(knn) > 1 {
			nearestPlayerComp, err := components.Player.Get(world, game.Players[knn[1].PersonaTag])
			left := tmpPlayer.Loc.First <= nearestPlayerComp.Loc.First

			if err != nil {
				return fmt.Errorf("Cardinal: error fetching player: %w", err)
			}

			if utils.Distance(nearestPlayerComp.Loc, prevLoc) <= attackRange {
				attackQueue = append(attackQueue, types.Triple[storage.EntityID, storage.EntityID, types.Triple[bool, string, string]]{
					First: game.Players[knn[1].PersonaTag],
					Second: tmpPlayer.Weapon,
					Third: types.Triple[bool, string, string]{left, tmpPlayer.PersonaTag, nearestPlayerComp.PersonaTag},
				})
			}
		}

		// moving players
		loc := utils.Move(tmpPlayer)

		point := &types.P{vector.V{prevLoc.First, prevLoc.Second}, personaTag}
		game.PlayerTree.Remove(point.P(), point.Equal)
		game.PlayerTree.Insert(&types.P{vector.V{loc.First, loc.Second}, personaTag})

		hitCoins := make([]types.Pair[storage.EntityID, types.Triple[float64, float64, int]], 0)
		hitHealth := make([]types.Pair[storage.EntityID, types.Triple[float64, float64, int]], 0)

		for i := int(math.Floor(prevLoc.First / game.GameParams.CSize)); i <= int(math.Floor(loc.First/game.GameParams.CSize)); i++ {
			for j := int(math.Floor(prevLoc.Second / game.GameParams.CSize)); j <= int(math.Floor(loc.Second/game.GameParams.CSize)); j++ {
				for coin, _ := range game.CoinMap[i][j] {
					if utils.CoinProjDist(prevLoc, loc, coin.Second) <= game.WorldConstants.PlayerRadius {
						hitCoins = append(hitCoins, coin)
					}
				}

				if tmpPlayer.Health < game.LevelHealth(tmpPlayer.Level) {
					for health, _ := range game.HealthMap[i][j] {
						if utils.CoinProjDist(prevLoc, loc, health.Second) <= game.WorldConstants.PlayerRadius {
							hitHealth = append(hitHealth, health)
						}
					}
				}
			}
		}

		extraCoins := 0

		for _, entityID := range hitCoins {
			if coinVal, err := utils.RemoveCoin(world, entityID); err != nil {
				return err
			} else {
				extraCoins += coinVal
			}
		}

		extraHealth := 0

		for _, entityID := range hitHealth {
			if healthVal, err := utils.RemoveHealth(world, entityID); err != nil {
				return err
			} else {
				extraHealth += healthVal
			}
		}

		// modifies player location and health
		components.Player.Update(world, id, func(comp components.PlayerComponent) components.PlayerComponent {
			comp.Loc = loc
			comp.Coins += extraCoins
			game.PlayerCoins[personaTag] = comp.Coins

			for game.LevelCoins(comp.Level) <= comp.Coins {
				comp.Coins -= game.LevelCoins(comp.Level)
				comp.Level++

				comp.Health = game.LevelHealth(comp.Level)
			}

			comp.Health += extraHealth
			if comp.Health > game.LevelHealth(comp.Level) {
				comp.Health = game.LevelHealth(comp.Level)
			}

			return comp
		})
	}

	for _, triple := range attackQueue {
		if err := utils.Attack(world, triple.First, triple.Second, triple.Third.First, triple.Third.Second, triple.Third.Third); err != nil {
			return err
		}
	}

	if float64(maxDepth) > 1 + float64(game.WorldConstants.BalanceFactor) * math.Log(float64(len(game.Players))) {
		game.PlayerTree.Balance()
	}

	return nil
}
