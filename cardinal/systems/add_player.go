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
	"github.com/rs/zerolog/log"

	"github.com/downflux/go-geometry/nd/vector"
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

		log.Debug().Msgf("Adding player with name: %s", tx.Name)

		// Create the player
		playerID, err := world.Create(components.Player)

		if err != nil {
			return fmt.Errorf("Error adding player to world: %w", err)
		}

		// Set the component to the correct values
		components.Player.Set(world, playerID, components.PlayerComponent{
			Name:   tx.Name,
			Coins:  tx.Coins,
			Weapon: types.Weapon(game.Melee), // This is the default weaponehn
		})

		// Add player to local PlayerTree
		playerComp, err := components.Player.Get(world, playerID)
		game.PlayerTree.Insert(&types.P{vector.V{playerComp.Loc.First, playerComp.Loc.Second}, playerComp.Name})
		log.Debug().Msgf("Created player with name", playerComp.Name)
	}

	return nil
}
