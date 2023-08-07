package main

import (
	"fmt"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/read"
	"github.com/argus-labs/new-game/systems"
	tx "github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/new-game/utils"
	"github.com/argus-labs/world-engine/cardinal/ecs/inmem"
	"github.com/argus-labs/world-engine/cardinal/server"
	//"github.com/rs/zerolog"
)

func main() {
	fmt.Println("Cardinal: SERVER HAS STARTED1")
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	//cfg := utils.GetConfig()

	// NOTE: Uses a Redis container
	// Best to use this for testing with Retool
	//world := cfg.World

	// NOTE: If you want to use an in-memory Redis, use this instead.
	// This is the easiest way to run Cardinal locally, but doen't work with Retool.
	// world := utils.NewInmemWorld()
	world := inmem.NewECSWorld()
	fmt.Println("Cardinal: SERVER HAS STARTED2")

	// Register components
	// NOTE: You must register your components here,
	// otherwise it will show an error when you try to use them in a system.
	utils.Must(world.RegisterComponents(
		components.Player,
		components.Coin,
		components.Health,
		components.Weapon,
	))
	fmt.Println("Cardinal: SERVER HAS STARTED3")

	// Register transactions
	// NOTE: You must register your transactions here,
	// otherwise it will show an error when you try to use them in a system.
	utils.Must(world.RegisterTransactions(
		tx.TxMovePlayer,
		tx.TxAddPlayer,
		tx.TxRemovePlayer,
		tx.TxSpawnCoins,
		tx.TxSpawnHealths,
	))
	fmt.Println("Cardinal: SERVER HAS STARTED4")

	// Register the reads
	utils.Must(world.RegisterReads(
		read.Constant,
		read.PlayerState,
		read.PlayerCoins,
		read.ReadAttacks,
		read.PlayerHealths,
		read.PlayerTotalCoins,
		read.ReadTick,
	))
	fmt.Println("Cardinal: SERVER HAS STARTED5")

	// Register the systems
	world.AddSystem(systems.AddPlayerSystem)
	fmt.Println("Cardinal: SERVER HAS STARTED6")
	world.AddSystem(systems.MoveSystem)
	fmt.Println("Cardinal: SERVER HAS STARTED7")
	world.AddSystem(systems.ProcessMovesSystem)
	fmt.Println("Cardinal: SERVER HAS STARTED8")
	world.AddSystem(systems.RemovePlayerSystem)
	fmt.Println("Cardinal: SERVER HAS STARTED9")
	world.AddSystem(systems.SpawnCoinsSystem)
	fmt.Println("Cardinal: SERVER HAS STARTED10")
	world.AddSystem(systems.SpawnHealthsSystem)
	fmt.Println("Cardinal: SERVER HAS STARTED11")

	// Load game state
	utils.Must(world.LoadGameState())
	fmt.Println("Cardinal: SERVER HAS STARTED12")

	// Setup some game settings
	gameSettings := types.Game{
		Dims: types.Pair[float64, float64]{
			First:  100,
			Second: 100,
		},
		CSize:   5,
		Players: []string{},
	}
	utils.InitializeGame(world, gameSettings)
	fmt.Println("Cardinal: SERVER HAS STARTED13")

	// Start game loop as a goroutine
	//go utils.GameLoop(world)

	// Register handlers
	h, err := server.NewHandler(world, server.DisableSignatureVerification())
	fmt.Println("Cardinal: SERVER HAS STARTED14")
	if err != nil {
		panic(err)
	}
	fmt.Println("Cardinal: SERVER HAS STARTED15")
	h.Serve("", "3333")
	fmt.Println("Cardinal: SERVER HAS STARTED16")
}
