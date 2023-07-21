package systems

import (
	"fmt"

	"github.com/argus-labs/new-game/components"
	msg2 "github.com/argus-labs/new-game/msg/query"
	msg1 "github.com/argus-labs/new-game/msg/tx"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func AddPlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	addPlayerTxs := msg1.TxAddPlayer.In(tq)

	for _, tx := range addPlayerTxs {
		if _, contains := Players[player.Name]; contains {
			utils.WriteError(w, "player name already exists", nil)
			return
		}

		playerID, err := world.Create(components.Player)
		if err != nil {
			return fmt.Errorf("Error adding player to world: %w", err)
		}

		m := msg2.QueryConstantMsg{
			ConstantLabel: "World",
		}

		components.Player.Set(world, playerID, components.PlayerComponent{
			Name:   tx.Name,
			Coins:  tx.Coins,
			Weapon: msg2.queryConstant(m),
		})

	}

	return nil
}
