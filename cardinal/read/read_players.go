package read

import (
	"errors"

	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/filter"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type PlayerPair struct {
	ID        storage.EntityID
	Component components.PlayerComponent
}

func ReadPlayers(world *ecs.World) []PlayerPair {
	// Create a slice to store the player pairs
	var playerPairs []PlayerPair

	ecs.NewQuery(filter.Exact(components.Player)).Each(world, func(id storage.EntityID) {
		// Get the player component for each player
		playerComp, err := components.Player.Get(world, id)
		if err != nil {
			return
		}

		// Create a PlayerPair and add it to the slice
		pair := PlayerPair{
			ID:        id,
			Component: playerComp,
		}
		playerPairs = append(playerPairs, pair)
	})

	// Return the slice containing player pairs
	return playerPairs
}

func GetPlayerByPersonaTag(world *ecs.World, personaTag string) (PlayerPair, error) {
	var player PlayerPair
	var err error
	ecs.NewQuery(filter.Exact(components.Player)).Each(world, func(id storage.EntityID) {
		playerComp, err := components.Player.Get(world, id)
		if err != nil {
			return
		}

		if playerComp.PersonaTag == personaTag {
			pair := PlayerPair{
				ID:        id,
				Component: playerComp,
			}
			player = pair
			err = nil
		}
	})
	if err != nil {
		err = errors.New("No player with the given PersonaTag exists.")
	}

	return player, err
}
