package msg

import (
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type RemovePlayer struct {
	Name string `json:"name"`
}

var TxRemovePlayer = ecs.NewTransactionType[RemovePlayer]("remove-player")
