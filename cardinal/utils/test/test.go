package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/argus-labs/new-game/component"
	tx "github.com/argus-labs/new-game/msg/tx"
	"github.com/argus-labs/new-game/system"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

// Miscellaneous test utilities
func ScaffoldTestWorld() *ecs.World {
	world := utils.NewInmemWorld()

	utils.Must(world.RegisterTransactions(
		tx.TxCreatePlayer,
	))

	world.AddSystem(system.PlayerSpawnerSystem)

	utils.Must(world.LoadGameState())

	return world
}

func CreatePlayer(world *ecs.World, tag string) (storage.EntityID, component.PlayerComponent) {
	entityId, _ := world.Create(component.Player)
	playerComp := component.PlayerComponent{
		Tag: tag,
	}
	component.Player.Set(world, entityId, playerComp)
	return entityId, playerComp
}

func DecodeMsg[T any](r *http.Request, msg *T) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	// Try to decode the request body into struct T.
	// Errors if msg has unknown fields.
	err := decoder.Decode(&msg)
	if err != nil {
		return err
	}

	// Check that all fields are present
	fields := reflect.ValueOf(msg).Elem()
	for i := 0; i < fields.NumField(); i++ {
		if fields.Field(i).IsZero() {
			return errors.New("some msg field(s) are missing")
		}
	}

	return nil
}
