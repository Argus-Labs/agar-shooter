package read

import (
	"encoding/json"

	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/world-engine/cardinal/ecs"
)

type ReadAttacksMsg struct {}

var ReadAttacks = ecs.NewReadType[ReadAttacksMsg]("attacks", readAttacks)

func readAttacks(world *ecs.World, m []byte) ([]byte, error) {
	returnMsg, err := json.Marshal(game.Attacks)

	return returnMsg, err
}
