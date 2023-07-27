package main

import (
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/read"
	"github.com/argus-labs/new-game/systems"
	tx "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/types"
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
		tx.TxAddPlayer,
		tx.TxRemovePlayer,
		tx.TxSpawnCoins,
	))

	// Register the reads
	utils.Must(world.RegisterReads(
		read.Constant,
		read.PlayerState,
		read.PlayersStatus,
		read.PlayerCoins,
		read.ReadAttacks,
	))

	// Register the systems
	world.AddSystem(systems.MoveSystem)
	world.AddSystem(systems.ProcessMovesSystem)
	world.AddSystem(systems.AddPlayerSystem)
	world.AddSystem(systems.RemovePlayerSystem)
	world.AddSystem(systems.SpawnCoinsSystem)

	// Load game state
	utils.Must(world.LoadGameState())

	// Setup some game settings
	gameSettings := types.Game{
		Dims: types.Pair[float64, float64]{
			First:  100,
			Second: 100,
		},
		CSize:   5,
		Players: []string{},
	}
	game.InitializeGame(world, gameSettings)

	// Start game loop as a goroutine
	go utils.GameLoop(world)

	// Register handlers
	h, err := server.NewHandler(world, server.DisableSignatureVerification())
	if err != nil {
		panic(err)
	}
	h.Serve("", cfg.CardinalPort)

	// DONE:
	//{"games/move", handleMakeMove},
	//{"games/push", handlePlayerPush}
	//{"games/pop", handlePlayerPop},
	//{"games/state", getPlayerState},
	//{"games/status", getPlayerStatus},
	//{"games/coins", getPlayerCoins},
	//{"games/create", createGame}
	//{"games/tick", tig},
	//{"games/offload", checkExtraction},
	//{"games/attacks", recentAttacks},
	//{"games/testaddhealth", testAddHealth},
}