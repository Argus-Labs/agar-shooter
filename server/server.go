package main

import (
	"fmt"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

func init() {
	World.RegisterComponents(ItemMapComp, PlayerMapComp, PlayerComp)
	MoveTx.SetID(0)
	//World.RegisterTransactions(MoveTx)
}

//func callback(cb func() error) {
//	cb()
//	time.Sleep(time.Second/tickRate)
//	callback(cb)
//}

// adds update Systems
func processMoves(World *ecs.World, q *ecs.TransactionQueue) error {// adjusts player directions based on their movement
	moveMap := make(map[string] Move)

	for _, move := range MoveTx.In(q) {
		moveMap[move.Player] = move
	}

	for name, move := range moveMap {
		PlayerComp.Update(World, Players[name], func(comp PlayerComponent) PlayerComponent {// modifies player direction struct
			diff := func(a, b bool) int {
				if a == b { return 0 }
				if a && !b { return 1 }
				return -1
			}

			comp.Dir.Face = Pair[int,int]{diff(move.Right, move.Left), diff(move.Up, move.Down)}// adjusts move direction
			return comp
		})
	}

	return nil
}

func makeMoves(World *ecs.World, q *ecs.TransactionQueue) error {// moves player based on the coin-speed; TODO: check collisions
	for playerName, id := range Players {
		tmpPlayer, err := PlayerComp.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		loc := Pair[int,int]{prevLoc.First + (10 * tmpPlayer.Dir.Face.First)/(1 + tmpPlayer.Coins), prevLoc.Second + (10 * tmpPlayer.Dir.Face.Second)/(1 + tmpPlayer.Coins)}// change speed function

		PlayerComp.Update(World, Players[playerName], func(comp PlayerComponent) PlayerComponent{// modifies player location
			comp.Loc = loc
			return comp
		})

		PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent{// moves player in map
			delete(comp.Players[Pair[int,int]{prevLoc.First/GameParams.CSize, prevLoc.Second/GameParams.CSize}], Pair[storage.EntityID, Pair[int,int]]{id, prevLoc})
			comp.Players[Pair[int,int]{loc.First/GameParams.CSize, loc.Second/GameParams.CSize}][Pair[storage.EntityID,Pair[int,int]]{id, loc}] = pewp

			return comp
		})
	}
	return nil
}


func HandlePlayerPush(player ModPlayer) error {
	playerID, err := World.Create(PlayerComp)// creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Melee, Pair[int,int]{25,25}, Direction{90, Pair[int,int]{0,0}}})// default player
	PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent {// adds a player to the board
		playercomp, err := PlayerComp.Get(World, Players[player.Name])

		if err != nil {
			fmt.Errorf("Error getting location with callback function: %w", err)
			return comp
		}

		newPlayer := Pair[storage.EntityID, Pair[int,int]]{Players[player.Name], playercomp.Loc}
		comp.Players[Pair[int,int]{25/GameParams.CSize,25/GameParams.CSize}][newPlayer] = pewp

		return comp
	})

	return nil
}

func HandlePlayerPop(player ModPlayer) error {
	//err := World.Remove(Players[player.Name])

	//if err != nil {
	//	return err
	//}

	PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent{//removes a player from the board
		playercomp, err := PlayerComp.Get(World, Players[player.Name])

		if err != nil {
			fmt.Errorf("error removing player: %w", err)
			return comp
		}

		oldPlayer := Pair[storage.EntityID, Pair[int,int]]{Players[player.Name], playercomp.Loc}
		delete(comp.Players[Pair[int,int]{playercomp.Loc.First/GameParams.CSize, playercomp.Loc.Second/GameParams.CSize}], oldPlayer)

		return comp
	})

	delete(Players, player.Name)

	return nil
}

func TickTock() {// testing function used to make the game tick
	World.Tick()
}

func GetPlayerState(player ModPlayer) (PlayerComponent, error) {// testing function used in place of broadcast to get state of players
	comp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		return PlayerComponent{}, fmt.Errorf("Player fetch error: %w", err)
	}

	return comp, nil
}

func HandleMakeMove(move Move) {
	MoveTx.AddToQueue(World, move)// adds "move" transaction to World transaction queue
}
