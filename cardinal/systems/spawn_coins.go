package systems

import (
	"errors"
	msg "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func SpawnCoinsSystem(world *ecs.World, tq *ecs.TransactionQueue) error {

	// For every tx, run SpawnCoins
	for _, _ = range msg.TxMovePlayer.In(tq) {
		if err := utils.SpawnCoins(world); err != nil {
			return errors.New("SpawnCoinSystem: Failed to run SpawnCoins function")
		}
	}

	return nil
}
