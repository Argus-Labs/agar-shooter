package msg

import "github.com/argus-labs/world-engine/cardinal/ecs"

type MovePlayerMsg struct {
	TargetPlayerTag		    string  `json:"target_player_tag"`
	PlayerID				string  `json:"playerID"`
	Up						bool    `json:"up"`
	Down					bool    `json:"down"`
	Left					bool    `json:"left"`
	Right					bool    `json:"right"`
	Input_sequence_number	int		`json:"input_sequence_number"`
	Delta					float64	`json:"delta"`
}

var TxMovePlayer = ecs.NewTransactionType[MovePlayerMsg]("move-player")
