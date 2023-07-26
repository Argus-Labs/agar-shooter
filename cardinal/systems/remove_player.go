package systems

import (
	tx "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func RemovePlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	removePlayerTxs := tx.TxRemovePlayer.In(tq)

	for _, tx := range removePlayerTxs {
		if _
	}
}
