package msg

import (
	"net/http"

	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type AddPlayerMsg struct {
	Name  string `json:"name"`
	Coins int    `json:"coins"`
}

var TxAddPlayer = ecs.NewTransactionType[AddPlayerMsg]("add-player")

// NOTE: We are going to be abstracting away this in the future, but for now
// you have to copy and paste this for each transaction type.
func (h *TxHandler) AddPlayer(w http.ResponseWriter, r *http.Request) {
	var msg AddPlayerMsg
	err := utils.DecodeMsg[AddPlayerMsg](r, &msg)
	if err != nil {
		utils.WriteError(w, "unable to decode add player tx", err)
		return
	}
	TxAddPlayer.AddToQueue(h.World, msg)
	utils.WriteResult(w, "ok")
}
