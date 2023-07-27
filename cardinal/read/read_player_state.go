package read

import (
	"encoding/json"
	"errors"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
)

type ReadPlayerStateMsg struct {
	PlayerName string `json:"player_name"`
}

var PlayerState = ecs.NewReadType[ReadPlayerStateMsg]("player-state", readPlayerState)

func readPlayerState(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var msg ReadPlayerStateMsg
	err := json.Unmarshal(m, &msg)
	if err != nil {
		return nil, err
	}

	// Check that the player exists
	var foundPlayer bool = false
	var foundPlayerID storage.EntityID
	players := ReadPlayers(world)
	for _, player := range players {
		if player.Component.Name == msg.PlayerName {
			foundPlayer = true
			foundPlayerID = player.ID
		}
	}
	if foundPlayer == false {
		log.Error().Msg("ReadPlayerState: Player with given name not found.")
		return nil, errors.New("ReadPlayerState: Player with given name not found")
	}

	// Get the Player's Component
	comp, err := components.Player.Get(world, foundPlayerID)
	if err != nil {
		log.Error().Msg("ReadPlayerState: Player component not found")
	}

	// Return the component as bytes
	var returnMsg []byte
	returnMsg, err = json.Marshal(comp)

	return returnMsg, nil
}
