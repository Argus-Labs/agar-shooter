package main

import (
	"fmt"
	"math"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

// world Systems
func processMoves(World *ecs.World, q *ecs.TransactionQueue) error {// adjusts player directions based on their movement
	moveMap := make(map[string] Move)

	for _, move := range MoveTx.In(q) {
		_, contains := moveMap[move.Player]
		if !contains || moveMap[move.Player].PacketNum < move.PacketNum{
			moveMap[move.Player] = move
		}
	}

	for name, move := range moveMap {
		naMe, contains := Players[name]
		
		if !contains {
			return fmt.Errorf("Cardinal: unregistered player attempting to move")
		}

		PlayerComp.Update(World, naMe, func(comp PlayerComponent) PlayerComponent {// modifies player direction struct
			diff := func(a, b bool) float64 {
				if a == b { return 0 }
				if a && !b { return 1 }
				return -1
			}

			comp.Dir.Face = Pair[float64,float64]{diff(move.Right, move.Left), diff(move.Up, move.Down)}// adjusts move direction
			comp.MoveNum = move.PacketNum		

			return comp
		})
	}

	return nil
}

func makeMoves(World *ecs.World, q *ecs.TransactionQueue) error {// moves player based on the coin-speed
	for playerName, id := range Players {
		tmpPlayer, err := PlayerComp.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		bound := func(x float64, y float64) (float64, float64){
			return math.Min(float64(GameParams.Dims.First), math.Max(0, x)), math.Min(float64(GameParams.Dims.Second), math.Max(0, y))
		}

		x, y := bound(prevLoc.First + (10 * tmpPlayer.Dir.Face.First)/float64(1 + tmpPlayer.Coins), prevLoc.Second + (10 * tmpPlayer.Dir.Face.Second)/float64(1 + tmpPlayer.Coins))
		
		loc := Pair[float64, float64]{x,y}// change speed function

		PlayerComp.Update(World, Players[playerName], func(comp PlayerComponent) PlayerComponent{// modifies player location
			comp.Loc = loc
			return comp
		})

		PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent{// moves player in map
			delete(comp.Players[Pair[int,int]{int(math.Floor(prevLoc.First/GameParams.CSize)), int(math.Floor(prevLoc.Second/GameParams.CSize))}], Pair[storage.EntityID, Pair[float64,float64]]{id, prevLoc})
			comp.Players[Pair[int,int]{int(math.Floor(loc.First/GameParams.CSize)), int(math.Floor(loc.Second/GameParams.CSize))}][Pair[storage.EntityID,Pair[float64,float64]]{id, loc}] = pewp

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

	PlayerComp.Set(World, Players[player.Name], PlayerComponent{player.Name, 100, 0, Melee, Pair[float64,float64]{25,25}, Direction{90, Pair[float64,float64]{0,0}}, 0})// default player
	PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent {// adds a player to the board
		playercomp, err := PlayerComp.Get(World, Players[player.Name])

		if err != nil {
			fmt.Errorf("Error getting location with callback function: %w", err)
			return comp
		}

		newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
		comp.Players[Pair[int,int]{25/int(GameParams.CSize),25/int(GameParams.CSize)}][newPlayer] = pewp

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

		oldPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[player.Name], playercomp.Loc}
		delete(comp.Players[Pair[int,int]{int(math.Floor(playercomp.Loc.First/GameParams.CSize)), int(math.Floor(playercomp.Loc.Second/GameParams.CSize))}], oldPlayer)

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
