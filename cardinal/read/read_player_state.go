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
	PlayerName string `json:"player_persona"`
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
	if _, contains := game.Players[msg.PlayerName]; !contains {
		return nil, fmt.Errorf("ReadPlayerState: Player with name %s not found", msg.PlayerName)
	}

	// Get the Player's Component
	comp, err := components.Player.Get(world, game.Players[msg.PlayerName])
	if err != nil {
		fmt.Errorf("ReadPlayerState: Player component not found")
	}

	// Return the component as bytes
	//log.Debug().Msgf("read-player-state: PlayerLocation is %v", comp.Loc)
	var returnMsg []byte
	returnMsg, err = json.Marshal(comp.Simplify())

	return returnMsg, nil
}
