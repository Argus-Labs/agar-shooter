package read

import (
	"context"
	"encoding/json"

	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type ReadTickMsg struct {}

var ReadTick = ecs.NewReadType[ReadTickMsg]("tick", readTick)

func readTick(world *ecs.World, m []byte) ([]byte, error) {
	world.Tick(context.Background())
	returnMsg, err := json.Marshal(world.CurrentTick())

	return returnMsg, err
}
