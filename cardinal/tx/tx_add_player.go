package msg

import (
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type AddPlayerMsg struct {
	PersonaTag string `json:"persona_tag"`
	Coins      int    `json:"coins"`
}

var TxAddPlayer = ecs.NewTransactionType[AddPlayerMsg]("add-player")
