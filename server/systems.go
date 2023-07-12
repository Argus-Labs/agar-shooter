package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

func diff(a, b bool) float64 {
	if a == b { return 0 }
	if a && !b { return 1 }
	return -1
}

// world Systems
func processMoves(World *ecs.World, q *ecs.TransactionQueue) error {// adjusts player directions based on their movement
	moveMap := make(map[string] []Move)

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

			for key, _ := range Players{
				str += " " + key
			}

			return fmt.Errorf("Cardinal: unregistered player attempting to move " + str)
		}

		var dir Pair[float64, float64]

		for _, move := range moveList {
			moove := Pair[float64,float64]{diff(move.Right, move.Left), diff(move.Up, move.Down)}
			norm := math.Max(1, math.Sqrt(moove.First*moove.First + moove.Second*moove.Second))

			dir = Pair[float64, float64]{dir.First + move.Delta*moove.First/norm, dir.Second + move.Delta*moove.Second/norm}
		}

		lastMove := Pair[float64, float64]{diff(moveList[len(moveList)-1].Right, moveList[len(moveList)-1].Left), diff(moveList[len(moveList)-1].Up, moveList[len(moveList)-1].Down)}

		PlayerComp.Update(World, entityID, func(comp PlayerComponent) PlayerComponent {// modifies player direction struct
			comp.Dir = dir// adjusts move directions
			comp.MoveNum = moveList[len(moveList)-1].Input_sequence_number
			comp.LastMove = lastMove
			if lastMove.First != 0 {
				comp.IsRight = lastMove.First > 0
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

func bound(x float64, y float64) Pair[float64, float64]{
	return Pair[float64, float64]{math.Min(float64(GameParams.Dims.First), math.Max(0, x)), math.Min(float64(GameParams.Dims.Second), math.Max(0, y))}
}

func distance(loc1, loc2 Mult) float64 {// returns distance between two coins
	return math.Sqrt(math.Pow(loc1.getFirst() - loc2.getFirst(), 2) + math.Pow(loc1.getSecond() - loc2.getSecond(), 2))
}

func move(tmpPlayer PlayerComponent) Pair[float64, float64] {// change speed function
	dir := tmpPlayer.Dir
	coins := tmpPlayer.Coins
	return bound(tmpPlayer.Loc.First + (sped * dir.First * math.Exp(-0.01*float64(coins))), tmpPlayer.Loc.Second + (sped * dir.Second * math.Exp(-0.01*float64(coins))))
}

func CoinProjDist(start, end, coin Pair[float64, float64]) float64 {// closest distance the coin is from the player obtained by checking the orthogonal projection of the coin with the segment defined by [start,end] TODO: write testcase for finding this value
	vec := Pair[float64, float64]{end.First-start.First, end.Second-start.Second}
	coeff := (vec.First*coin.First + vec.Second*coin.Second)/(vec.First*vec.First + vec.Second*vec.Second)
	proj := Pair[float64, float64]{coeff*vec.First + start.First, coeff*vec.Second + start.Second}
	ortho := Pair[float64, float64]{coin.First - proj.First, coin.Second-proj.Second}

	if proj.First*vec.First + proj.Second*vec.Second < 0 {// if the coin is outside of the span of the orthogonal, return the distance to the closest endpoint
		return math.Min(math.Sqrt(math.Pow(coin.First - start.First, 2) + math.Pow(coin.Second - start.Second, 2)), math.Sqrt(math.Pow(coin.First - end.First, 2) + math.Pow(coin.Second - end.Second, 2)))
	}

	return math.Sqrt(ortho.First*ortho.First + ortho.Second*ortho.Second)
}

func attack(id storage.EntityID, weapon Weapon, hurt bool) error {// attack a player
	kill := false
	coins := false
	var loc Pair[float64, float64]
	var name string

	if err := PlayerComp.Update(World, id, func(comp PlayerComponent) PlayerComponent{// modifies player location
		if !hurt && comp.Coins > 0 {
			comp.Coins--
			coins = true
		} else{
			comp.Health -= Weapons[weapon].Attack
		}
		kill = comp.Health <= 0
		name = comp.Name
		loc = comp.Loc

		return comp
	}); err != nil {
		return nil
	}

	if coins {
		randfloat := rand.Float64()
		loc = bound(loc.First + 3*math.Cos(randfloat*2*math.Pi), loc.Second + 3*math.Sin(randfloat*2*math.Pi))
		coinID, err := World.Create(CoinComp)

		if err != nil {
			return fmt.Errorf("Coin creation failed: %w", err)
		}

		CoinMap[GetCell(loc)][Pair[storage.EntityID, Triple[float64, float64, int]]{coinID, Triple[float64,float64,int]{loc.First,loc.Second,1}}] = pewp
		CoinComp.Set(World, coinID, CoinComponent{loc, 1})
	}

	if kill {// removes player from map if they die
		if err := HandlePlayerPop(ModPlayer{name}); err != nil {
			return err
		}
	}

	return nil
}

func makeMoves(World *ecs.World, q *ecs.TransactionQueue) error {// moves player based on the coin-speed
	attackQueue := make([]Triple[storage.EntityID, Weapon, bool],0)
	Attacks = make([]AttackTriple, 0)

	for playerName, id := range Players {
		tmpPlayer, err := PlayerComp.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		// attacking players; each player attacks the closest player TODO: change targetting system later

		var (
			minID storage.EntityID
			minDistance float64
			closestPlayerName string
			left bool
		)

		assigned := false

		for _, closestPlayerID := range Players {
			if closestPlayerID != id {
				closestPlayer, err := PlayerComp.Get(World, closestPlayerID)
				if err != nil {
					return err
				}
			
				dist := distance(closestPlayer.Loc, prevLoc)

				if !assigned || minDistance > dist {
					minID = closestPlayerID
					minDistance = dist
					closestPlayerName = closestPlayer.Name
					assigned = true
					left = closestPlayer.Loc.First <= tmpPlayer.Loc.First
				}
			}
		}

		if assigned && minDistance <= Weapons[tmpPlayer.Weapon].Range {
			attackQueue = append(attackQueue, Triple[storage.EntityID, Weapon, bool]{minID, tmpPlayer.Weapon, left == tmpPlayer.IsRight})
			attackVal := 0
			if left == tmpPlayer.IsRight && tmpPlayer.Loc > 0 {
				attackVal = -1
			} else {
				attackVal = Weapons[tmpPlayer.Weapon].Attack
			}
			Attacks = append(Attacks, AttackTriple{playerName, closestPlayerName, attackVal})
		}

		// moving players

		loc := move(tmpPlayer)

		delete(PlayerMap[Pair[int,int]{int(math.Floor(prevLoc.First/GameParams.CSize)), int(math.Floor(prevLoc.Second/GameParams.CSize))}], Pair[storage.EntityID, Pair[float64,float64]]{id, prevLoc})
		PlayerMap[Pair[int,int]{int(math.Floor(loc.First/GameParams.CSize)), int(math.Floor(loc.Second/GameParams.CSize))}][Pair[storage.EntityID,Pair[float64,float64]]{id, loc}] = pewp
		
		hitCoins := make([]Pair[storage.EntityID, Triple[float64,float64,int]], 0)

		for i := int(math.Floor(prevLoc.First/GameParams.CSize)); i <= int(math.Floor(loc.First/GameParams.CSize)); i++ {
			for j := int(math.Floor(prevLoc.Second/GameParams.CSize)); j <= int(math.Floor(loc.Second/GameParams.CSize)); j++ {
				for coin, _ := range CoinMap[Pair[int, int]{i,j}] {
					if distance(prevLoc, coin.Second) <= PlayerRadius || distance(loc, coin.Second) <= PlayerRadius {//CoinProjDist(prevLoc, loc, coin.Second) <= PlayerRadius {
						hitCoins = append(hitCoins, coin)
					}
				}
			}
		}

		extraCoins := 0

		for _, entityID := range hitCoins {
			coin, err := CoinComp.Get(World, entityID.First)

			if err != nil {
				fmt.Errorf("Cardinal: could not get coin entity")
			}

			extraCoins += coin.Val
			delete(CoinMap[Pair[int,int]{int(math.Floor(entityID.Second.First/GameParams.CSize)),int(math.Floor(entityID.Second.Second/GameParams.CSize))}], entityID)

			if err := World.Remove(entityID.First); err != nil {
				return err
			}
		}

		PlayerComp.Update(World, Players[playerName], func(comp PlayerComponent) PlayerComponent{// modifies player location
			comp.Loc = loc
			comp.Coins += extraCoins
			
			return comp
		})
	}

	for _, triple := range attackQueue {
		if err := attack(triple.First, triple.Second, triple.Third); err != nil {
			return err
		}
	}

	return nil
}
