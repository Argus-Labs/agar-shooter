package read

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"time"
)

type ReadTickMsg struct{}

var ReadTick = ecs.NewReadType[ReadTickMsg]("tick", readTick)

func readTick(world *ecs.World, m []byte) ([]byte, error) {
	//fmt.Printf("Current tick before: %d\n", world.CurrentTick())

	startTime := time.Now()
	world.Tick(context.Background())
	endTime := time.Now()
	//fmt.Printf("Current tick after: %d\n", world.CurrentTick())
	fmt.Printf("Cardinal tick completion time is %d", endTime.Sub(startTime).Milliseconds())
	returnMsg, err := json.Marshal(world.CurrentTick())

	return returnMsg, err
}
