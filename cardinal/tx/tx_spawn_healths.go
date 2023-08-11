package msg

import "github.com/argus-labs/world-engine/cardinal/ecs"

type SpawnHealthsMsg struct{}

var TxSpawnHealths = ecs.NewTransactionType[SpawnHealthsMsg]("spawn-healths")
