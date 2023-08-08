package read

import (
	"encoding/json"
	"fmt"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
)

type ReadPlayerStateMsg struct {
	PlayerPersona string `json:"player_persona"`
}

var PlayerState = ecs.NewReadType[ReadPlayerStateMsg]("player-state", readPlayerState)

func readPlayerState(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var msg ReadPlayerStateMsg
	err := json.Unmarshal(m, &msg)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("ReadPlayerStateMsg: %+v", msg)

	// Check that the player exists
	var foundPlayer bool = false
	var foundPlayerID storage.EntityID
	players := ReadPlayers(world)
	for _, player := range players {
		if player.Component.PersonaTag == msg.PlayerPersona {
			foundPlayer = true
			foundPlayerID = player.ID
		}
	}
	if foundPlayer == false {
		return nil, fmt.Errorf("ReadPlayerState: Player with name %s not found", msg.PlayerPersona)
	}

	// Get the Player's Component
	comp, err := components.Player.Get(world, foundPlayerID)
	if err != nil {
		log.Error().Msg("ReadPlayerState: Player component not found")
	}

	// Return the component as bytes
	log.Debug().Msgf("read-player-state: PlayerLocation is %v", comp.Loc)
	var returnMsg []byte
	returnMsg, err = json.Marshal(comp.Simplify())

	return returnMsg, nil
}
