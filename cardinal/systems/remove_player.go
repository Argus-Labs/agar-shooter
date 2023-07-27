package systems

import (
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/read"
	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func RemovePlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	removePlayerTxs := transactions.TxRemovePlayer.In(tq)
	var err error = nil
	playerList := read.ReadPlayers(world)

	for _, tx := range removePlayerTxs {
		err = game.RemovePlayer(world, tx.Name, playerList)
	}
	return err
}
