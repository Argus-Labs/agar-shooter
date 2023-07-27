package systems

import (
	"errors"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/read"
	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
	"math"
)

func RemovePlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	removePlayerTxs := transactions.TxRemovePlayer.In(tq)
	var err error = nil
	players := read.ReadPlayers(world)

	for _, tx := range removePlayerTxs {

		// Check that the player exists
		var playerFound bool = false
		for _, player := range players {
			if player.Component.Name == tx.Name {
				playerFound = true
			}
		}
		if playerFound == false {
			log.Error().Msg("player name already exists")
			return errors.New("RemovePlayerSystem: Player does not exist.")
		}

		// Get the player id and component
		player, err := read.GetPlayerByName(world, tx.Name)
		if err != nil {
			return err
		}

		// Remove the player from the World
		if err := world.Remove(player.ID); err != nil {
			log.Error().Msg("RemovePlayerSystem: Error removing player.")
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

			peep := game.bound(player.Component.Loc.First+rad*math.Cos(2*math.Pi*float64(start)/float64(tot)), player.Component.Loc.Second+rad*math.Sin(2*math.Pi*float64(start)/float64(tot)))
			newCoins = append(newCoins, types.Triple[float64, float64, int]{peep.First, peep.Second, addCoins})
			start++
		}

		for _, coin := range newCoins {
			if _, err := game.AddCoin(world, coin); err != nil {
				return err
			}
		}

		// Delete the player from the local PlayerMap
		oldPlayer := types.Pair[storage.EntityID, types.Pair[float64, float64]]{player.ID, player.Component.Loc}
		delete(game.PlayerMap[types.GetCell(player.Component.Loc)], oldPlayer)

		delete(game.Players, player.Component.Name)
	}
	return err
}
