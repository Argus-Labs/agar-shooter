package read

import (
	"encoding/json"
	"fmt"

	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

var PlayerTotalCoins = ecs.NewReadType[ReadPlayerCoinsMsg]("player-total-coins", readPlayerTotalCoins)

func getTotalCoins(playerName string) int {
	return game.PlayerCoins[playerName]
}

func readPlayerTotalCoins(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var msg ReadPlayerCoinsMsg
	err := json.Unmarshal(m, &msg)
	if err != nil {
		return nil, err
	}

	// Check that the player exists
	if _, contains := game.Players[msg.PlayerName]; !contains {
		return nil, fmt.Errorf("ReadPlayerCoins: Player with given name not found")
	}

	totalCoins := getTotalCoins(msg.PlayerName)

	// Return the component as bytes
	returnMsg, err := json.Marshal(totalCoins)

	return returnMsg, err
}
