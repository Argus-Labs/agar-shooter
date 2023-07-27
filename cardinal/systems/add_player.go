package systems

import (
	"errors"
	"fmt"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/read"
	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
)

func AddPlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	addPlayerTxs := transactions.TxAddPlayer.In(tq)
	players := read.ReadPlayers(world)

	for _, tx := range addPlayerTxs {

		// check if player already exists; don't do anything
		for _, player := range players {
			if player.Component.Name == tx.Name {
				log.Error().Msg("player name already exists")
				return errors.New("AddPlayerSystem: Player name already exists.")
			}
		}

		// Create the player
		playerID, err := world.Create(components.Player)
		if err != nil {
			return fmt.Errorf("Error adding player to world: %w", err)
		}

		// Set the component to the correct values
		components.Player.Set(world, playerID, components.PlayerComponent{
			Name:   tx.Name,
			Coins:  tx.Coins,
			Weapon: types.Weapon(game.Melee), // This is the default weapon
		})

		// Add player to local PlayerMap
		playerComp, err := components.Player.Get(world, playerID)
		newPlayer := types.Pair[storage.EntityID, types.Pair[float64, float64]]{playerID, playerComp.Loc}
		game.PlayerMap[types.GetCell(playerComp.Loc)][newPlayer] = types.Pewp
	}

	return nil
}
