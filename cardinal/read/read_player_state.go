package read

import (
	"encoding/json"
	"fmt"

	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type ReadPlayerStateMsg struct {
	PersonaTag string `json:"player_persona"`
}

var PlayerState = ecs.NewReadType[ReadPlayerStateMsg]("player-state", readPlayerState)

func readPlayerState(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var pkg types.Package[ReadPlayerStateMsg]//ReadPlayerStateMsg
	err := json.Unmarshal(m, &pkg)
	if err != nil {
		return nil, err
	}

	msg := pkg.Body

	// Check that the player exists
	if _, contains := game.Players[msg.PersonaTag]; !contains {
		return nil, fmt.Errorf("ReadPlayerState: Player with PersonaTag %s not found", msg.PersonaTag)
	}

	// Get the Player's Component
	comp, err := components.Player.Get(world, game.Players[msg.PersonaTag])
	if err != nil {
		fmt.Errorf("ReadPlayerState: Player component not found")
	}

	// Return the component as bytes
	var returnMsg []byte
	returnMsg, err = json.Marshal(comp.Simplify())

	return returnMsg, nil
}
