package systems

import (
	"github.com/argus-labs/new-game/read"
	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/rs/zerolog/log"
)

func RemovePlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	removePlayerTxs := transactions.TxRemovePlayer.In(tq)
	var err error = nil
	playerList := read.ReadPlayers(world)

	for _, tx := range removePlayerTxs {
		log.Debug().Msgf("Removing player with name: %s", tx.Name)
		err = utils.RemovePlayer(world, tx.Name, playerList)
	}
	return err
}
