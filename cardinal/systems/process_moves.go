package systems

import (
	"fmt"
	"math"

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

// world Systems
func processMoves(World *ecs.World, q *ecs.TransactionQueue) error { // adjusts player directions based on their movement
	moveMap := make(map[string][]Move)

	for _, move := range MoveTx.In(q) {
		if _, contains := moveMap[move.PlayerID]; !contains {
			/*
				pcomp, err := PlayerComp.Get(World, Players[move.PlayerID])

				if err != nil {
					return err
				}

				if pcomp.MoveNum != move.Input_sequence_number - 1{
					fmt.Printf("Difference in input sequence number is not 1; received sequence number %i after sequence number %i.",move.Input_sequence_number,pcomp.MoveNum)
					return nil
				}
			*/

			moveMap[move.PlayerID] = []Move{move}
		} else {
			/*
				if num := moveMap[move.PlayerID][len(moveMap[move.PlayerID])-1].Input_sequence_number;move.Input_sequence_number != num + 1 {
					fmt.Printf("Difference in input sequence number is not 1; received sequence number %i after sequence number %i.",move.Input_sequence_number,num)
					return nil
				}
			*/
			moveMap[move.PlayerID] = append(moveMap[move.PlayerID], move)
		}
	}

	for name, moveList := range moveMap {
		entityID, contains := Players[name]

		if !contains {
			str := ""

			for key, _ := range Players {
				str += " " + key
			}

			return fmt.Errorf("Cardinal: unregistered player attempting to move " + str)
		}

		var dir types.Pair[float64, float64]
		isRight := false

		for _, move := range moveList {
			moove := types.Pair[float64, float64]{diff(move.Right, move.Left), diff(move.Up, move.Down)}
			norm := math.Max(1, math.Sqrt(moove.First*moove.First+moove.Second*moove.Second))

			dir = types.Pair[float64, float64]{dir.First + move.Delta*moove.First/norm, dir.Second + move.Delta*moove.Second/norm}
			if moove.First != 0 {
				isRight = moove.First > 0
			}
		}

		lastMove := types.Pair[float64, float64]{diff(moveList[len(moveList)-1].Right, moveList[len(moveList)-1].Left), diff(moveList[len(moveList)-1].Up, moveList[len(moveList)-1].Down)}

		PlayerComp.Update(World, entityID, func(comp PlayerComponent) PlayerComponent { // modifies player direction struct
			comp.Dir = dir // adjusts move directions
			comp.MoveNum = moveList[len(moveList)-1].Input_sequence_number
			comp.LastMove = lastMove
			if lastMove.First != 0 {
				comp.IsRight = isRight
			}

			return comp
		})
	}

	for player, entityID := range Players {
		_, contains := moveMap[player]
		if contains {
			continue
		}

		PlayerComp.Update(World, entityID, func(comp PlayerComponent) PlayerComponent {
			comp.Dir = comp.LastMove

			return comp
		})
	}

	return nil
}
