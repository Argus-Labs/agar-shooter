package read

import (
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/filter"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type PlayerPair struct {
	ID     storage.EntityID
	Player components.PlayerComponent
}

func ReadPlayers(world *ecs.World) []PlayerPair {
	// Create a slice to store the player pairs
	var playerPairs []PlayerPair

	ecs.NewQuery(filter.Exact(components.Player)).Each(world, func(id storage.EntityID) {
		// Get the player component for each player
		player, err := components.Player.Get(world, id)
		if err != nil {
			return
		}

		// Create a PlayerPair and add it to the slice
		pair := PlayerPair{
			ID:     id,
			Player: player,
		}
		playerPairs = append(playerPairs, pair)
	})

	// Return the slice containing player pairs
	return playerPairs
}
