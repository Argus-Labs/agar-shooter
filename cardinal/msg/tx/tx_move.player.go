package msg

import (
	"net/http"

	"github.com/argus-labs/new-game/utils"
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

var TxMovePlayer = ecs.NewTransactionType[MovePlayerMsg]()

// NOTE: We are going to be abstracting away this in the future, but for now
// you have to copy and paste this for each transaction type.
func (h *TxHandler) MovePlayer(w http.ResponseWriter, r *http.Request) {
	var msg MovePlayerMsg
	err := utils.DecodeMsg[MovePlayerMsg](r, &msg)
	if err != nil {
		utils.WriteError(w, "unable to decode move player tx", err)
		return
	}
	TxMovePlayer.AddToQueue(h.World, msg)
	utils.WriteResult(w, "ok")
}
