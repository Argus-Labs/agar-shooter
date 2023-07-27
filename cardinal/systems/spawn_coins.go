package systems

import (
	"errors"
	"github.com/argus-labs/new-game/game"
	msg "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func SpawnCoinsSystem(world *ecs.World, tq *ecs.TransactionQueue) error {

	// For every tx, run SpawnCoins
	for _, _ = range msg.TxMovePlayer.In(tq) {
		if err := game.SpawnCoins(world); err != nil {
			return errors.New("SpawnCoinSystem: Failed to run SpawnCoins function")
		}
	}

	return nil
}
