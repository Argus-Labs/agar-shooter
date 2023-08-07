package systems

import (
	"fmt"

	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/rs/zerolog/log"
)

func AddPlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	addPlayerTxs := transactions.TxAddPlayer.In(tq)

	for _, tx := range addPlayerTxs {
		log.Debug().Msgf("Adding player with PersonaTag: %s", tx.PersonaTag)
		if err := utils.AddPlayer(world, tx.PersonaTag, tx.Coins); err != nil {
			return fmt.Errorf("Cardinal: error adding player:", err)
		}
	}

	return nil
}
