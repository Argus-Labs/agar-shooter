package systems

import (
	"github.com/argus-labs/new-game/read"
	transactions "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/rs/zerolog/log"
)

func RemovePlayerSystem(world *ecs.World, tq *ecs.TransactionQueue) error {
	removePlayerTxs := transactions.TxRemovePlayer.In(tq)
	var err error = nil
	playerList := read.ReadPlayers(world)
	sum := 0
	for _, tx := range removePlayerTxs {
		sum += 1
		log.Debug().Msgf("Removing player with PersonaTag: %s", tx.PersonaTag)
		err = utils.RemovePlayer(world, tx.PersonaTag, playerList)
	}
	//fmt.Printf("RemovePlayer TXs: %d\n", sum)
	return err
}
