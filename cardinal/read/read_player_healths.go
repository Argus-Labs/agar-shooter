package read

import (
	"encoding/json"
	"errors"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
	"math"
)

type ReadPlayerHealthsMsg struct {
	PlayerName string `json:"player_name"`
}

var PlayerHealths = ecs.NewReadType[ReadPlayerHealthsMsg]("player-health", readPlayerHealths)

func getNearbyHealths(playerComp components.PlayerComponent) []types.NearbyHealth {
	healths := make([]types.NearbyHealth, 0)

	for i := math.Max(0, math.Floor((playerComp.Loc.First-game.ClientView.First/2)/game.GameParams.CSize)); i <= math.Min(float64(game.Width), math.Ceil((playerComp.Loc.First+game.ClientView.First/2)/game.GameParams.CSize)); i++ {
		for j := math.Max(0, math.Floor((playerComp.Loc.Second-game.ClientView.Second/2)/game.GameParams.CSize)); j <= math.Min(float64(game.Height), math.Ceil((playerComp.Loc.Second+game.ClientView.Second/2)/game.GameParams.CSize)); j++ {
			for health, _ := range game.HealthMap[types.Pair[int, int]{First: int(i), Second: int(j)}] {
				healths = append(healths, types.NearbyHealth{health.Second.First, health.Second.Second, health.Second.Third})
			}
		}
	}

	return healths
}

func readPlayerHealths(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var msg ReadPlayerHealthsMsg
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
		return nil, errors.New("ReadPlayerHealths: Player with given name not found")
	}

	// Get the Player's Component
	comp, err := components.Player.Get(world, foundPlayerID)
	if err != nil {
		log.Error().Msg("ReadPlayerHealths: Player component not found")
	}

	var nearbyHealths []types.NearbyHealth
	nearbyHealths = getNearbyHealths(comp)

	// Return the component as bytes
	returnMsg, err := json.Marshal(nearbyHealths)

	return returnMsg, err
}