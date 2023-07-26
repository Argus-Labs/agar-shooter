package main

import (
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/read"
	"github.com/argus-labs/new-game/systems"
	tx "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/server"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	cfg := utils.GetConfig()

	// NOTE: Uses a Redis container
	// Best to use this for testing with Retool
	world := cfg.World

	// NOTE: If you want to use an in-memory Redis, use this instead.
	// This is the easiest way to run Cardinal locally, but doen't work with Retool.
	// world := utils.NewInmemWorld()

	// Register components
	// NOTE: You must register your components here,
	// otherwise it will show an error when you try to use them in a system.
	utils.Must(world.RegisterComponents(
		components.Player,
		components.Coin,
		components.Health,
		components.Weapon,
	))

	// Register transactions
	// NOTE: You must register your transactions here,
	// otherwise it will show an error when you try to use them in a system.
	utils.Must(world.RegisterTransactions(
		tx.TxMovePlayer,
	))

	utils.Must(world.RegisterReads(
		read.Constant))

	world.AddSystem(systems.MoveSystem)
	// world.AddSystem(systems.ProcessMovesSystem)

	// Load game state
	utils.Must(world.LoadGameState())

	// Start game loop as a goroutine
	go utils.GameLoop(world)

	// Register handlers
	h, err := server.NewHandler(world, server.DisableSignatureVerification())
	if err != nil {
		panic(err)
	}
	h.Serve("", cfg.CardinalPort)

	// DONE:
	//{"tx_move_player", tx.MovePlayer}

	// TODO: NOT DONE
	//{"games/push", handlePlayerPush}
	//{"games/pop", handlePlayerPop},
	//{"games/move", handleMakeMove},
	//{"games/state", getPlayerState},
	//{"games/status", getPlayerStatus},
	//{"games/coins", getPlayerCoins},
	//{"games/tick", tig},
	//{"games/create", createGame},
	//{"games/offload", checkExtraction},
	//{"games/attacks", recentAttacks},
	//{"games/testaddhealth", testAddHealth},
}
