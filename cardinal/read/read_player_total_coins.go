package read

import (
	"encoding/json"
	"fmt"

	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/new-game/types"
)

type ReadPlayerTotalCoinsMsg struct {
	PersonaTag string `json:"player_persona"`
}

var PlayerTotalCoins = ecs.NewReadType[ReadPlayerTotalCoinsMsg]("player-total-coins", readPlayerTotalCoins)

func getTotalCoins(personaTag string) int {
	return game.PlayerCoins[personaTag]
}

func readPlayerTotalCoins(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var pkg types.Package[ReadPlayerTotalCoinsMsg]
	err := json.Unmarshal(m, &pkg)
	if err != nil {
		return nil, err
	}

	msg := pkg.Body

	// Check that the player exists
	if _, contains := game.Players[msg.PersonaTag]; !contains {
		return nil, fmt.Errorf("ReadPlayerCoins: Player with given PersonaTag not found")
	}

	totalCoins := getTotalCoins(msg.PersonaTag)

	// Return the component as bytes
	returnMsg, err := json.Marshal(totalCoins)

	return returnMsg, err
}
