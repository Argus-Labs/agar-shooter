package msg

import (
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type AddPlayerMsg struct {
	Name  string `json:"name"`
	Coins int    `json:"coins"`
}

var TxAddPlayer = ecs.NewTransactionType[AddPlayerMsg]("add-player")
