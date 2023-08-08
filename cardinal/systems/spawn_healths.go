package systems

import (
	"fmt"

	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func SpawnHealthsSystem(world *ecs.World, tq *ecs.TransactionQueue) error {

	if err := utils.SpawnHealths(world); err != nil {
		return fmt.Errorf("SpawnHealthSystem: Failed to run SpawnHealths function", err)
	}

	return nil
}
