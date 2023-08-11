package systems

import (
	"fmt"

	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

func AddPlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	addPlayerTxs := transactions.TxAddPlayer.In(tq)
	sum := 0
	for _, tx := range addPlayerTxs {
		sum += 1
		fmt.Printf("Adding player with PersonaTag: %s", tx.PersonaTag)
		if err := utils.AddPlayer(world, tx.PersonaTag, tx.Coins); err != nil {
			return fmt.Errorf("Cardinal: error adding player:", err)
		}
	}

	fmt.Printf("AddPlayerTxs: %d\n", sum)
	return nil
}
