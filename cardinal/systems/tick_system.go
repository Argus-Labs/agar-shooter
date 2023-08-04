package systems

import (
	"context"
	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/rs/zerolog/log"
)

func TickSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	tickTxs := transactions.TxTick.In(tq)
	var err error = nil

	for _, _ = range tickTxs {
		log.Debug().Msgf("Executing tick system")
		if err = world.Tick(context.Background()); err != nil {
			return err
		}
	}
	return err
}
