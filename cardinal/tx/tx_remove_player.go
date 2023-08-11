package msg

import "github.com/argus-labs/world-engine/cardinal/ecs"

type RemovePlayerMsg struct {
	PersonaTag string `json:"name"`
}

var TxRemovePlayer = ecs.NewTransactionType[RemovePlayerMsg]("remove-player")
