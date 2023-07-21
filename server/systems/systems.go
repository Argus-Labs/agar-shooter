package systems

import (
	"fmt"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/component"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/game"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/types"
	"math"
	"math/rand"

	"github.com/argus-labs/world-engine/cardinal/ecs"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"
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
func ProcessMoves(World *ecs.World, q *ecs.TransactionQueue) error { // adjusts player directions based on their movement
	moveMap := make(map[string][]types.Move)

	for _, move := range game.MoveTx.In(q) {
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

			moveMap[move.PlayerID] = []types.Move{move}
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
		entityID, contains := game.Players[name]

		if !contains {
			str := ""

			for key, _ := range game.Players {
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

		game.PlayerComp.Update(World, entityID, func(comp component.PlayerComponent) component.PlayerComponent { // modifies player direction struct
			comp.Dir = dir // adjusts move directions
			comp.MoveNum = moveList[len(moveList)-1].Input_sequence_number
			comp.LastMove = lastMove
			if lastMove.First != 0 {
				comp.IsRight = isRight
			}

			return comp
		})
	}

	for player, entityID := range game.Players {
		_, contains := moveMap[player]
		if contains {
			continue
		}

		game.PlayerComp.Update(World, entityID, func(comp component.PlayerComponent) component.PlayerComponent {
			comp.Dir = comp.LastMove

			return comp
		})
	}

	return nil
}

func bound(x float64, y float64) types.Pair[float64, float64] {
	return types.Pair[float64, float64]{math.Min(float64(game.GameParams.Dims.First), math.Max(0, x)), math.Min(float64(game.GameParams.Dims.Second), math.Max(0, y))}
}

func distance(loc1, loc2 types.Mult) float64 { // returns distance between two coins
	return math.Sqrt(math.Pow(loc1.getFirst()-loc2.getFirst(), 2) + math.Pow(loc1.getSecond()-loc2.getSecond(), 2))
}

func move(tmpPlayer component.PlayerComponent) types.Pair[float64, float64] { // change speed function
	dir := tmpPlayer.Dir
	coins := tmpPlayer.Coins
	return bound(tmpPlayer.Loc.First+(game.sped*dir.First*math.Exp(-0.01*float64(coins))), tmpPlayer.Loc.Second+(game.sped*dir.Second*math.Exp(-0.01*float64(coins))))
}

func CoinProjDist(start, end types.Pair[float64, float64], coin types.Triple[float64, float64, int]) float64 { // closest distance the coin is from the player obtained by checking the orthogonal projection of the coin with the segment defined by [start,end]
	vec := types.Pair[float64, float64]{end.First - start.First, end.Second - start.Second}
	coin = types.Triple[float64, float64, int]{coin.First - start.First, coin.Second - start.Second, 0}
	coeff := (vec.First*coin.First + vec.Second*coin.Second) / (vec.First*vec.First + vec.Second*vec.Second)
	proj := types.Pair[float64, float64]{coeff*vec.First + start.First, coeff*vec.Second + start.Second}
	ortho := types.Pair[float64, float64]{coin.First - proj.First, coin.Second - proj.Second}

	if proj.First*vec.First+proj.Second*vec.Second < 0 || proj.First*proj.First+proj.Second*proj.Second > vec.First*vec.First+vec.Second*vec.Second { // if the coin is outside the span of the segment, return the distance to the closest endpoint
		return math.Sqrt(math.Min(coin.First*coin.First+coin.Second*coin.Second, (coin.First-vec.First)*(coin.First-vec.First)+(coin.Second-vec.Second)*(coin.Second-vec.Second)))
	} else {
		return math.Sqrt(ortho.First*ortho.First + ortho.Second*ortho.Second)
	}
}

func attack(id storage.EntityID, weapon types.Weapon, left bool, attacker, defender string) error { // attack a player
	kill := false
	coins := false
	var loc types.Pair[float64, float64]
	var name string

	if err := game.PlayerComp.Update(game.World, id, func(comp component.PlayerComponent) component.PlayerComponent { // modifies player location
		if left == comp.IsRight && comp.Coins > 0 {
			comp.Coins--
			coins = true
		} else {
			comp.Health -= game.Weapons[weapon].Attack
		}
		kill = comp.Health <= 0
		name = comp.Name
		loc = comp.Loc

		return comp
	}); err != nil {
		return nil
	}

	if coins {
		randfloat := rand.Float64() * 2 * math.Pi
		loc = bound(loc.First+3*math.Cos(randfloat), loc.Second+3*math.Sin(randfloat))

		if _, err := game.AddCoin(types.Triple[float64, float64, int]{loc.First, loc.Second, 1}); err != nil {
			return err
		}

		game.Attacks = append(game.Attacks, types.AttackTriple{attacker, defender, -1})
	} else { // adds attack to display queue if it was executed
		game.Attacks = append(game.Attacks, types.AttackTriple{attacker, defender, game.Weapons[weapon].Attack})
	}

	if kill { // removes player from map if they die
		if err := game.HandlePlayerPop(types.ModPlayer{name}); err != nil {
			return err
		}
	}

	return nil
}

func MakeMoves(World *ecs.World, q *ecs.TransactionQueue) error { // moves player based on the coin-speed
	attackQueue := make([]types.Triple[storage.EntityID, types.Weapon, types.Triple[bool, string, string]], 0)
	game.Attacks = make([]types.AttackTriple, 0)
	maxDepth := 0

	for playerName, id := range game.Players {
		tmpPlayer, err := game.PlayerComp.Get(World, id)

		if err != nil {
			return err
		}

		prevLoc := tmpPlayer.Loc

		// attacking players; each player attacks the closest player as determined by kdtree TODO: change targetting system later
		depth := 0
		knn := kd.KNN[*types.P](game.PlayerTree, vector.V{prevLoc.First, prevLoc.Second}, 2, func(q *types.P) bool {
			depth++
			return true
		})

		if maxDepth < depth {
			maxDepth = depth
		}

		if len(knn) > 1 {
			nearestPlayerComp, err := game.PlayerComp.Get(World, game.Players[knn[1].Name])
			left := tmpPlayer.Loc.First <= nearestPlayerComp.Loc.First
			if err != nil {
				return fmt.Errorf("Cardinal: error fetching player: %w", err)
			}
			if distance(nearestPlayerComp.Loc, prevLoc) <= game.Weapons[tmpPlayer.Weapon].Range {
				attackQueue = append(attackQueue, types.Triple[storage.EntityID, types.Weapon, types.Triple[bool, string, string]]{game.Players[knn[1].Name], tmpPlayer.Weapon, types.Triple[bool, string, string]{left, playerName, nearestPlayerComp.Name}})
			}
		}

		// moving players --- this is the only place players move, so modifying player tree values must only occur here (outside of inserts and deletes)
		loc := move(tmpPlayer)

		point := &types.P{vector.V{prevLoc.First, prevLoc.Second}, playerName}
		game.PlayerTree.Remove(point.P(), point.Equal)
		game.PlayerTree.Insert(&types.P{vector.V{loc.First, loc.Second}, playerName})

		// collects all hit coins
		hitCoins := make([]types.Pair[storage.EntityID, types.Triple[float64, float64, int]], 0)

		for i := int(math.Floor(prevLoc.First / game.GameParams.CSize)); i <= int(math.Floor(loc.First/game.GameParams.CSize)); i++ {
			for j := int(math.Floor(prevLoc.Second / game.GameParams.CSize)); j <= int(math.Floor(loc.Second/game.GameParams.CSize)); j++ {
				for coin, _ := range game.CoinMap[types.Pair[int, int]{i, j}] {
					if CoinProjDist(prevLoc, loc, coin.Second) <= game.PlayerRadius {
						hitCoins = append(hitCoins, coin)
					}
				}
			}
		}

		extraCoins := 0

		for _, entityID := range hitCoins {
			if coinVal, err := game.RemoveCoin(entityID); err != nil {
				return err
			} else {
				extraCoins += coinVal
			}

		}

		game.PlayerComp.Update(World, game.Players[playerName], func(comp component.PlayerComponent) component.PlayerComponent { // modifies player location
			comp.Loc = loc
			comp.Coins += extraCoins
			if game.PlayerMaxCoins[playerName] < comp.Coins {
				game.PlayerMaxCoins[playerName] = comp.Coins
			}

			return comp
		})
	}

	for _, triple := range attackQueue {
		if err := attack(triple.First, triple.Second, triple.Third.First, triple.Third.Second, triple.Third.Third); err != nil {
			return err
		}
	}

	if float64(maxDepth) > 1+game.balanceFactor*math.Log(float64(len(game.Players))) {
		game.PlayerTree.Balance()
	}

	return nil
}
