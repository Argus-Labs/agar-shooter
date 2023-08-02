package systems

import (
	"fmt"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/read"
	"github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/rs/zerolog/log"
	"math"
)

func diff(a, b bool) float64 {
	if a == b {
		return 0
	}
	if a && !b {
		return 1
	}
	return -1
}

// adjusts player directions based on their movement
func MoveSystem(world *ecs.World, q *ecs.TransactionQueue) error {
	// playerId -> Move Directions Struct mapping
	moveMap := make(map[string][]msg.MovePlayerMsg)
	log.Debug().Msgf("Entered MoveSystem, world.CurrentTick: %d", world.CurrentTick())
	// Build the moveMap from the txQueue
	for key, move := range msg.TxMovePlayer.In(q) {
		log.Debug().Msgf("Found a TX number %d for the current tick", key)
		if _, contains := moveMap[move.PlayerID]; !contains {
			moveMap[move.PlayerID] = []msg.MovePlayerMsg{move}
		} else {
			moveMap[move.PlayerID] = append(moveMap[move.PlayerID], move)
		}
	}

	for playerID, playerMoveList := range moveMap {
		contains := false
		var entityID storage.EntityID
		players := read.ReadPlayers(world)
		for _, player := range players {
			if player.Component.PersonaTag == playerID {
				contains = true
				entityID = player.ID
			}
		}

		// Check if the player making the move is registered
		if !contains {
			//str := ""
			//
			//for key, _ := range game.Players {
			//	str += " " + key
			//}
			//continue
			return fmt.Errorf("Cardinal: unregistered player attempting to move ")
		}

		var dir types.Pair[float64, float64]
		isRight := false

		for _, move := range playerMoveList {
			moove := types.Pair[float64, float64]{
				First:  diff(move.Right, move.Left), // Calculate the difference between right and left movement
				Second: diff(move.Up, move.Down),    // Calculate the difference between up and down movement
			}
			norm := math.Max(1, math.Sqrt(moove.First*moove.First+moove.Second*moove.Second)) // Calculate the magnitude of the movement vector

			dir = types.Pair[float64, float64]{
				First:  dir.First + move.Delta*moove.First/norm,   // Update the current horizontal direction based on movement inputs
				Second: dir.Second + move.Delta*moove.Second/norm, // Update the current vertical direction based on movement inputs
			}
			if moove.First != 0 {
				isRight = moove.First > 0 // Determine the dominant horizontal movement direction
			}
		}

		lastMove := types.Pair[float64, float64]{
			First:  diff(playerMoveList[len(playerMoveList)-1].Right, playerMoveList[len(playerMoveList)-1].Left), // Calculate the difference between the latest right and left movement
			Second: diff(playerMoveList[len(playerMoveList)-1].Up, playerMoveList[len(playerMoveList)-1].Down),    // Calculate the difference between the latest up and down movement
		}

		// Update the player's direction in their PlayerComponent in Cardinal
		components.Player.Update(world, entityID, func(comp components.PlayerComponent) components.PlayerComponent {
			log.Debug().Msgf("tx-move-player: Updating player direction with the following attributes")
			log.Debug().Msgf("dir: %v", dir)
			log.Debug().Msgf("MoveNum: %d", playerMoveList[len(playerMoveList)-1].InputSequenceNumber)
			log.Debug().Msgf("LastMove: %v", lastMove)
			comp.Dir = dir                                                           // Adjust the player's move directions
			comp.MoveNum = playerMoveList[len(playerMoveList)-1].InputSequenceNumber // Set the player's latest input sequence number
			comp.LastMove = lastMove                                                 // Update the player's last movement
			if lastMove.First != 0 {
				comp.IsRight = isRight // Set the dominant horizontal movement direction
			}

			return comp
		})

	}
	players := read.ReadPlayers(world)
	for _, player := range players {
		_, contains := moveMap[player.Component.PersonaTag]
		if contains {
			continue
		}

		components.Player.Update(world, player.ID, func(comp components.PlayerComponent) components.PlayerComponent {
			comp.Dir = comp.LastMove
			return comp
		})
	}

	return nil
}
