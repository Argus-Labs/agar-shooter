package systems

import (
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/read"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
	"math"
)

// moves player based on the coin-speed
func ProcessMovesSystem(world *ecs.World, q *ecs.TransactionQueue) error {
	attackQueue := make([]types.Triple[storage.EntityID, types.Weapon, types.Triple[bool, string, string]], 0)
	game.Attacks = make([]types.AttackTriple, 0)
	log.Debug().Msgf("Entered ProcessMovesSystem, world.CurrentTick: %d", world.CurrentTick())

	players := read.ReadPlayers(world)
	// playerName, playerId
	for _, player := range players {
		player1Id := player.ID
		player1Name := player.Component.Name
		tmpPlayer, err := components.Player.Get(world, player1Id)

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

		for _, player := range players {
			if player.ID != player1Id {
				closestPlayer, err := components.Player.Get(world, player.ID)
				if err != nil {
					return err
				}

				dist := utils.Distance(closestPlayer.Loc, prevLoc)

				if !assigned || minDistance > dist {
					minID = player.ID
					minDistance = dist
					closestPlayerName = closestPlayer.Name
					assigned = true
					left = tmpPlayer.Loc.First <= closestPlayer.Loc.First
				}
			}
		}

		if assigned && minDistance <= game.WorldConstants.Weapons[tmpPlayer.Weapon].Range {
			log.Debug().Msgf("Player with name %s attacks player with name %s", player1Name, closestPlayerName)
			attackQueue = append(attackQueue, types.Triple[storage.EntityID, types.Weapon, types.Triple[bool, string, string]]{First: minID, Second: tmpPlayer.Weapon, Third: types.Triple[bool, string, string]{left, player1Name, closestPlayerName}})
		}

		// moving players
		loc := utils.Move(tmpPlayer)

		delete(game.PlayerMap[utils.GetCell(prevLoc)], types.Pair[storage.EntityID, types.Pair[float64, float64]]{First: player1Id, Second: prevLoc})
		game.PlayerMap[utils.GetCell(loc)][types.Pair[storage.EntityID, types.Pair[float64, float64]]{First: player1Id, Second: loc}] = types.Pewp

		hitCoins := make([]types.Pair[storage.EntityID, types.Triple[float64, float64, int]], 0)

		for i := int(math.Floor(prevLoc.First / game.GameParams.CSize)); i <= int(math.Floor(loc.First/game.GameParams.CSize)); i++ {
			for j := int(math.Floor(prevLoc.Second / game.GameParams.CSize)); j <= int(math.Floor(loc.Second/game.GameParams.CSize)); j++ {
				for coin, _ := range game.CoinMap[types.Pair[int, int]{First: i, Second: j}] {
					if utils.CoinProjDist(prevLoc, loc, coin.Second) <= game.WorldConstants.PlayerRadius {
						hitCoins = append(hitCoins, coin)
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

		// modifies player location
		components.Player.Update(world, player1Id, func(comp components.PlayerComponent) components.PlayerComponent {
			log.Debug().Msgf("Updating player location to: %v", loc)
			comp.Loc = loc
			comp.Coins += extraCoins

			return comp
		})
	}

	for _, triple := range attackQueue {
		if err := utils.Attack(world, triple.First, triple.Second, triple.Third.First, triple.Third.Second, triple.Third.Third); err != nil {
			return err
		}
	}

	return nil
}
