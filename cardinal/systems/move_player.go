package systems

import (
	"fmt"
	"math"

	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/game"
	"github.com/argus-labs/new-game/tx"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs"
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
	sum := 0
	// Build the moveMap from the txQueue
	for _, move := range msg.TxMovePlayer.In(q) {
		sum += 1
		if _, contains := moveMap[move.PlayerID]; !contains {
			// The commented code is used to verify that no packets are lost when the client is sending inputs to the server, but rather than erroring out when a packet is unrecoverable, we choose to ignore it so we don't stop the game just because a packet was not sent
			/*
				pcomp, err := components.Player.Get(world, game.Players[move.PlayerID])

				if err != nil {
					return err
				}

				if pcomp.MoveNum != move.Input_sequence_number -1 {
					fmt.Println("Difference in input sequence number is not 1; received sequence number", move.Input_sequence_number, "after sequence number", pcomp.MoveNum)
				}
			*/

			moveMap[move.PlayerID] = []msg.MovePlayerMsg{move}
		} else {
			/*
				if num := moveMap[move.PlayerID][len(moveMap[move.PlayerID]) - 1].Input_sequence_number; move.Input_sequence_number != num + 1 {
					fmt.Println("Difference in input sequence number is not 1; received sequence number", move.Input_sequence_number, "after sequence number", num)
					return nil
				}
			*/

			moveMap[move.PlayerID] = append(moveMap[move.PlayerID], move)
		}
	}

	for personaTag, moveList := range moveMap {
		entityID, contains := game.Players[personaTag]

		// Check if the player making the move is registered
		if !contains {
			str := ""

			for key, _ := range game.Players {
				str += " " + key
			}

			return fmt.Errorf("Cardinal: unregistered player attempting to move " + str)
		}

		var dir types.Pair[float64, float64]
		isRight, nonZero := false, false

		for _, move := range moveList {
			moove := types.Pair[float64, float64]{
				First:  diff(move.Right, move.Left), // Calculate the difference between right and left movement
				Second: diff(move.Up, move.Down),    // Calculate the difference between up and down movement
			}
			norm := math.Max(1, math.Sqrt(moove.First*moove.First+moove.Second*moove.Second)) // Calculate the magnitude of the movement vector

			dir = types.Pair[float64, float64]{
				First:  dir.First + (move.Delta*moove.First)/norm,   // Update the current horizontal direction based on movement inputs
				Second: dir.Second + (move.Delta*moove.Second)/norm, // Update the current vertical direction based on movement inputs
			}
			if moove.First != 0 {
				nonZero = true
				isRight = moove.First > 0 // Determine the dominant horizontal movement direction
			}
		}

		lastMove := types.Pair[float64, float64]{
			First:  diff(moveList[len(moveList)-1].Right, moveList[len(moveList)-1].Left), // Calculate the difference between the latest right and left movement
			Second: diff(moveList[len(moveList)-1].Up, moveList[len(moveList)-1].Down),    // Calculate the difference between the latest up and down movement
		}

		// Update the player's direction in their PlayerComponent in Cardinal
		components.Player.Update(world, entityID, func(comp components.PlayerComponent) components.PlayerComponent {
			comp.Dir = dir                                                 // Adjust the player's move directions
			comp.MoveNum = moveList[len(moveList)-1].Input_sequence_number // Set the player's latest input sequence number
			comp.LastMove = lastMove                                       // Update the player's last movement
			if nonZero {
				comp.IsRight = isRight // Set the dominant horizontal movement direction
			}

			return comp
		})
	}

	for player, entityID := range game.Players {
		_, contains := moveMap[player]
		if contains {
			continue
		}

		components.Player.Update(world, entityID, func(comp components.PlayerComponent) components.PlayerComponent {
			comp.Dir = comp.LastMove

			return comp
		})
	}
	//fmt.Printf("MoveTxs: %d\n", sum)
	return nil
}
