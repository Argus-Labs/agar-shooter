package msg

import (
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type TickMsg struct{}

var TxTick = ecs.NewTransactionType[TickMsg]("tick")
