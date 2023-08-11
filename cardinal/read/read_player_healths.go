package read

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/rs/zerolog/log"

	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type ReadPlayerHealthsMsg struct {
	PersonaTag string `json:"player_persona"`
}

var PlayerHealths = ecs.NewReadType[ReadPlayerHealthsMsg]("player-health", readPlayerHealths)

func getNearbyHealths(playerComp components.PlayerComponent) []types.NearbyHealth {
	healths := make([]types.NearbyHealth, 0)

	for i := math.Max(0, math.Floor((playerComp.Loc.First-game.ClientView.First/2)/game.GameParams.CSize)); i <= math.Min(float64(game.Width), math.Ceil((playerComp.Loc.First+game.ClientView.First/2)/game.GameParams.CSize)); i++ {
		for j := math.Max(0, math.Floor((playerComp.Loc.Second-game.ClientView.Second/2)/game.GameParams.CSize)); j <= math.Min(float64(game.Height), math.Ceil((playerComp.Loc.Second+game.ClientView.Second/2)/game.GameParams.CSize)); j++ {
			for health, _ := range game.HealthMap[int(i)][int(j)] {
				healths = append(healths, types.NearbyHealth{health.Second.First, health.Second.Second, health.Second.Third})
			}
		}
	}

	return healths
}

func readPlayerHealths(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var pkg types.Package[ReadPlayerHealthsMsg]
	err := json.Unmarshal(m, &pkg)
	if err != nil {
		return nil, err
	}
	
	msg := pkg.Body

	// Check that the player exists
	if _, contains := game.Players[msg.PersonaTag]; !contains {
		return nil, errors.New("ReadPlayerHealths: Player with given PersonaTag not found: " +  msg.PersonaTag)
	}

	// Get the Player's Component
	comp, err := components.Player.Get(world, game.Players[msg.PersonaTag])
	if err != nil {
		log.Error().Msg("ReadPlayerHealths: Player component not found")
	}

	var nearbyHealths []types.NearbyHealth
	nearbyHealths = getNearbyHealths(comp)

	// Return the component as bytes
	returnMsg, err := json.Marshal(nearbyHealths)

	return returnMsg, err
}
