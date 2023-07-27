package read

import (
	"encoding/json"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/filter"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type ReadPlayersStatusMsg struct{}

var PlayersStatus = ecs.NewReadType[ReadPlayersStatusMsg]("players-status", readPlayersStatus)

func readPlayersStatus(world *ecs.World, m []byte) ([]byte, error) {

	// Create mapping of Player Name to Player Location
	nameToLocation := make([]types.Pair[string, types.Pair[float64, float64]], 0)

	// Query all players, add name and location to the mapping
	ecs.NewQuery(filter.Exact(components.Player)).Each(world, func(id storage.EntityID) {
		// Get the player component for each player
		playerComp, err := components.Player.Get(world, id)
		if err != nil {
			return
		}

		nameToLocation = append(nameToLocation, types.Pair[string, types.Pair[float64, float64]]{
			First:  playerComp.Name,
			Second: playerComp.Loc,
		})
	})

	// Return the component as bytes
	var returnMsg []byte
	returnMsg, _ = json.Marshal(nameToLocation)

	return returnMsg, nil
}
