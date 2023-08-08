package systems

import (
	"fmt"

	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func SpawnCoinsSystem(world *ecs.World, tq *ecs.TransactionQueue) error {

	if err := utils.SpawnCoins(world); err != nil {
		return fmt.Errorf("SpawnCoinSystem: Failed to run SpawnCoins function", err)
	}

	return nil
}
