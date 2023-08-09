package read

import (
	"encoding/json"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/new-game/game"
)

type ReadGameParametersMsg struct {}

type GameParameterStruct struct {
	TickRate				int
	PlayerSpeed				int
	Width					int
	Height					int
	WeaponRadius			float64
	LevelCoinParameters		[]float64
	LevelHealthParameters	[]float64
}

var GameParameters = ecs.NewReadType[ReadGameParametersMsg]("game-parameters", readGameParameters)

func readGameParameters(world *ecs.World, m []byte) ([]byte, error) {
	returnMsg, err := json.Marshal(GameParameterStruct{1000/game.WorldConstants.TickRate, game.WorldConstants.PlayerSpeed, int(game.GameParams.Dims.First), int(game.GameParams.Dims.Second), game.WorldConstants.Weapons[game.DefaultWeapon].Range, game.LevelCoinParameters, game.LevelHealthParameters})

	return returnMsg, err
}
