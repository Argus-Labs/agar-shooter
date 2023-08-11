package msg

import "github.com/argus-labs/world-engine/cardinal/ecs"

type SpawnCoinsMsg struct{}

var TxSpawnCoins = ecs.NewTransactionType[SpawnCoinsMsg]("spawn-coins")
