package read

import (
	"encoding/json"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/types"
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

	log.Info().Msgf("ReadPlayerStateMsg: %+v", msg)

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
		// TODO: put these errors back in
		log.Error().Msgf("ReadPlayerState: Player with name %s not found", msg.PlayerName)
		tempPlayer := &components.PlayerComponent{
			Name:   msg.PlayerName,
			Health: 100,
			Coins:  0,
			Weapon: 0,
			Loc: types.Pair[float64, float64]{
				First:  -1,
				Second: -1,
			},
			Dir: types.Pair[float64, float64]{
				First:  0,
				Second: 0,
			},
			LastMove: types.Pair[float64, float64]{
				First:  0,
				Second: 0,
			},
			Extract: types.Pair[float64, float64]{
				First:  0,
				Second: 0,
			},
			IsRight: true,
			MoveNum: 0,
		}

		var returnMsg []byte
		returnMsg, err = json.Marshal(tempPlayer)
		return returnMsg, nil
		//return returnMsg, fmt.Errorf("ReadPlayerState: Player with name %s not found", msg.PlayerName)
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
