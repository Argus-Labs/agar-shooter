package read

import (
	"encoding/json"
	"fmt"

	"math"

	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/rs/zerolog/log"
)

type ReadPlayerCoinsMsg struct {
	PlayerName string `json:"player_name"`
}

var PlayerCoins = ecs.NewReadType[ReadPlayerCoinsMsg]("player-coins", readPlayerCoins)

func getNearbyCoins(playerComp components.PlayerComponent) []types.NearbyCoin {
	coins := make([]types.NearbyCoin, 0)

	for i := math.Max(0, math.Floor((playerComp.Loc.First-game.ClientView.First/2)/game.GameParams.CSize)); i <= math.Min(float64(game.Width), math.Ceil((playerComp.Loc.First+game.ClientView.First/2)/game.GameParams.CSize)); i++ {
		for j := math.Max(0, math.Floor((playerComp.Loc.Second-game.ClientView.Second/2)/game.GameParams.CSize)); j <= math.Min(float64(game.Height), math.Ceil((playerComp.Loc.Second+game.ClientView.Second/2)/game.GameParams.CSize)); j++ {
			for coin, _ := range game.CoinMap[types.Pair[int, int]{First: int(i), Second: int(j)}] {
				coins = append(coins, types.NearbyCoin{coin.Second.First, coin.Second.Second, coin.Second.Third})
			}
		}
	}

	return coins
}

func readPlayerCoins(world *ecs.World, m []byte) ([]byte, error) {

	// Read the msg data from bytes
	var msg ReadPlayerCoinsMsg
	err := json.Unmarshal(m, &msg)
	if err != nil {
		return nil, err
	}

	// Check that the player exists
	if _, contains := game.Players[msg.PlayerName]; !contains {
		return nil, fmt.Errorf("ReadPlayerCoins: Player with given name not found")
	}

	// Get the Player's Component
	comp, err := components.Player.Get(world, game.Players[msg.PlayerName])
	if err != nil {
		log.Error().Msg("ReadPlayerCoins: Player component not found")
	}

	var nearbyCoins []types.NearbyCoin
	nearbyCoins = getNearbyCoins(comp)

	// Return the component as bytes
	returnMsg, err := json.Marshal(nearbyCoins)

	return returnMsg, err
}
