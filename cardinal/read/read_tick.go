package read

import (
	"context"
	"encoding/json"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type ReadTickMsg struct{}

var ReadTick = ecs.NewReadType[ReadTickMsg]("tick", readTick)

func readTick(world *ecs.World, m []byte) ([]byte, error) {
	//fmt.Printf("Current tick before: %d\n", world.CurrentTick())

	//startTime := time.Now()
	world.Tick(context.Background())
	//endTime := time.Now()
	//fmt.Printf("Current tick after: %d\n", world.CurrentTick())
	//fmt.Printf("Time taken for tick to complete is %d", endTime.Sub(startTime).Milliseconds())
	returnMsg, err := json.Marshal(world.CurrentTick())

	return returnMsg, err
}
