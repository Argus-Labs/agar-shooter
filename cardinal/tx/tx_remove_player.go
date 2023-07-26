package msg

import (
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type RemovePlayer struct {
	Name  string `json:"name"`
	Coins int    `json:"coins"`
}

var TxRemovePlayer = ecs.NewTransactionType[RemovePlayer]("remove-player")
