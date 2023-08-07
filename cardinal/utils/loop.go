package utils

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/new-game/game"
)

func GameLoop(world *ecs.World) {
	log.Info().Msg("Game loop started")
	for range time.Tick(time.Duration(time.Millisecond.Nanoseconds() * int64(game.WorldConstants.TickRate))) {
		if err := world.Tick(context.Background()); err != nil {
			panic(err)
		}
	}
}
