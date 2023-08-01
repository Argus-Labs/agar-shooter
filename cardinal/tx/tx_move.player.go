package msg

import (
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type MovePlayerMsg struct {
	TargetPlayerTag     string  `json:"target_player_tag"`
	PlayerID            string  `json:"playerID"`
	Up                  bool    `json:"up"`
	Down                bool    `json:"down"`
	Left                bool    `json:"left"`
	Right               bool    `json:"right"`
	InputSequenceNumber int     `json:"input_sequence_number"`
	Delta               float64 `json:"delta"`
}

var TxMovePlayer = ecs.NewTransactionType[MovePlayerMsg]("move-player")
