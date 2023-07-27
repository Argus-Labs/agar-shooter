package systems

import (
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"math"
)

// moves player based on the coin-speed
func ProcessMovesSystem(world *ecs.World, q *ecs.TransactionQueue) error {
	attackQueue := make([]types.Triple[storage.EntityID, types.Weapon, types.Triple[bool, string, string]], 0)
	game.Attacks = make([]types.AttackTriple, 0)

	for playerName, id := range game.Players {
		tmpPlayer, err := components.Player.Get(world, id)

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
				closestPlayer, err := components.Player.Get(world, closestPlayerID)
				if err != nil {
					return err
				}

				dist := game.Distance(closestPlayer.Loc, prevLoc)

				if !assigned || minDistance > dist {
					minID = closestPlayerID
					minDistance = dist
					closestPlayerName = closestPlayer.Name
					assigned = true
					left = tmpPlayer.Loc.First <= closestPlayer.Loc.First
				}
			}
		}

		if assigned && minDistance <= game.WorldConstants.Weapons[tmpPlayer.Weapon].Range {
			attackQueue = append(attackQueue, types.Triple[storage.EntityID, types.Weapon, types.Triple[bool, string, string]]{First: minID, Second: tmpPlayer.Weapon, Third: types.Triple[bool, string, string]{left, playerName, closestPlayerName}})
		}

		// moving players

		loc := game.Move(tmpPlayer)

		delete(game.PlayerMap[types.GetCell(prevLoc)], types.Pair[storage.EntityID, types.Pair[float64, float64]]{First: id, Second: prevLoc})
		game.PlayerMap[types.GetCell(loc)][types.Pair[storage.EntityID, types.Pair[float64, float64]]{First: id, Second: loc}] = types.Pewp

		hitCoins := make([]types.Pair[storage.EntityID, types.Triple[float64, float64, int]], 0)

		for i := int(math.Floor(prevLoc.First / game.GameParams.CSize)); i <= int(math.Floor(loc.First/game.GameParams.CSize)); i++ {
			for j := int(math.Floor(prevLoc.Second / game.GameParams.CSize)); j <= int(math.Floor(loc.Second/game.GameParams.CSize)); j++ {
				for coin, _ := range game.CoinMap[types.Pair[int, int]{i, j}] {
					if game.CoinProjDist(prevLoc, loc, coin.Second) <= game.WorldConstants.PlayerRadius {
						hitCoins = append(hitCoins, coin)
					}
				}
			}
		}

		extraCoins := 0

		for _, entityID := range hitCoins {
			if coinVal, err := game.RemoveCoin(world, entityID); err != nil {
				return err
			} else {
				extraCoins += coinVal
			}

		}

		// modifies player location
		components.Player.Update(world, game.Players[playerName], func(comp components.PlayerComponent) components.PlayerComponent {
			comp.Loc = loc
			comp.Coins += extraCoins

			return comp
		})
	}

	for _, triple := range attackQueue {
		if err := game.Attack(world, triple.First, triple.Second, triple.Third.First, triple.Third.Second, triple.Third.Third); err != nil {
			return err
		}
	}

	return nil
}
