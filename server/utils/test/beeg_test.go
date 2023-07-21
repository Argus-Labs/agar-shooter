package utils

import (
	"fmt"
	"github.com/argus-labs/world-engine/game/sample_game_server/server"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/component"
	game2 "github.com/argus-labs/world-engine/game/sample_game_server/server/game"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/types"
	"gotest.tools/v3/assert"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
)

type Message struct {
	message []byte
}

func (m Message) Header() http.Header {
	return nil
}

func (m Message) Write(data []byte) (int, error) {
	m.message = data
	return 0, nil
}

func (m Message) WriteHeader(statusCode int) {

}

func TestPewp(t *testing.T) {
	// test game initialization
	const LENGTH = 1000
	const ATTACKPLAYERS = 100
	game := types.Game{types.Pair[float64, float64]{LENGTH, LENGTH}, 1, []string{}}

	var err error
	var player types.ModPlayer
	var contains bool
	var p component.PlayerComponent

	err = game2.CreateGame(game)

	if err != nil {
		fmt.Println("pewp")
		fmt.Println("%w", err)
	}

	// test adding player moves and making player move each tick
	testPlayer1, testPlayer2, testPlayer3 := types.ModPlayer{"a"}, types.ModPlayer{"b"}, types.ModPlayer{"c"}
	game2.PushPlayer(component.PlayerComponent{"a", 100, 0, game2.Dud, types.Pair[float64, float64]{250, 250}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{500, 500}, true, -1})
	game2.PushPlayer(component.PlayerComponent{"b", 100, 0, game2.Dud, types.Pair[float64, float64]{750, 750}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{250, 250}, true, -1})

	m := make(map[types.ModPlayer][]types.TestPlayer)

	testMove := func(players []types.ModPlayer) {
		for _, player = range players {
			_, contains = m[player]

			if !contains {
				m[player] = make([]types.TestPlayer, 0)
			}

			p, err = game2.GetPlayerState(player)
			fmt.Println(player, ": ", p)
			assert.NilError(t, err)
			m[player] = append(m[player], p.Testify())
		}
	}

	testMove([]types.ModPlayer{testPlayer1, testPlayer2})

	move := types.Move{"a", true, false, true, false, 0, 0.2} // up, down, left, right
	game2.HandleMakeMove(move)

	testMove([]types.ModPlayer{testPlayer1, testPlayer2})

	game2.TickTock()

	testMove([]types.ModPlayer{testPlayer1, testPlayer2})

	assert.DeepEqual(t, m[testPlayer1][0], m[testPlayer1][1])
	assert.DeepEqual(t, m[testPlayer2][0], m[testPlayer2][1])

	assert.Assert(t, m[testPlayer1][1] != m[testPlayer1][2])
	assert.DeepEqual(t, m[testPlayer2][1], m[testPlayer2][2])

	// test adding a player to the game
	_, err = game2.GetPlayerState(testPlayer3)
	assert.Assert(t, err != nil)

	fmt.Println("start push")
	game2.PushPlayer(component.PlayerComponent{"c", 100, 0, game2.Dud, types.Pair[float64, float64]{500, 500}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{750, 750}, true, -1})
	fmt.Println("end push")

	testMove([]types.ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	// test moving players after a new player has been added
	game2.TickTock()
	testMove([]types.ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	fmt.Println(m[testPlayer1])
	assert.Assert(t, m[testPlayer1][len(m[testPlayer1])-2] != m[testPlayer1][len(m[testPlayer1])-1])
	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2])-2], m[testPlayer2][len(m[testPlayer2])-1])
	assert.DeepEqual(t, m[testPlayer3][len(m[testPlayer3])-2], m[testPlayer3][len(m[testPlayer3])-1])

	newMove := types.Move{"c", false, true, false, true, 0, 0.2}
	game2.HandleMakeMove(newMove)
	game2.TickTock()

	testMove([]types.ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	assert.Assert(t, m[testPlayer1][len(m[testPlayer1])-2] != m[testPlayer1][len(m[testPlayer1])-1])
	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2])-2], m[testPlayer2][len(m[testPlayer2])-1])
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-2] != m[testPlayer3][len(m[testPlayer3])-1])

	// test moving players after a player has been removed
	testMove([]types.ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	fmt.Println("start pop")
	game2.HandlePlayerPop(testPlayer1)
	fmt.Println("end pop")

	_, err = game2.GetPlayerState(testPlayer1)
	assert.Assert(t, err != nil)

	game2.TickTock()

	testMove([]types.ModPlayer{testPlayer2, testPlayer3})

	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2])-2], m[testPlayer2][len(m[testPlayer2])-1])
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-2] != m[testPlayer3][len(m[testPlayer3])-1])

	// test that players do not go beyond boundaries
	p2X := m[testPlayer2][len(m[testPlayer2])-1].LocX
	p3Y := m[testPlayer3][len(m[testPlayer3])-1].LocY

	move1 := types.Move{"b", true, false, false, false, 0, 0.2} // up, down, left, right
	move2 := types.Move{"c", false, false, false, true, 1, 0.2}

	game2.HandleMakeMove(move1)
	game2.HandleMakeMove(move2)
	for i := 0; i < 5*LENGTH; i++ {
		game2.TickTock()
	}

	testMove([]types.ModPlayer{testPlayer2, testPlayer3})
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocX == p2X)
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-1].LocY == p3Y)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocY == LENGTH)
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-1].LocX == LENGTH)

	move3 := types.Move{"b", false, true, true, false, 1, 0.2} // up, down, left, right
	game2.HandleMakeMove(move3)

	for i := 0; i < 5*LENGTH; i++ {
		game2.TickTock()
	}

	testMove([]types.ModPlayer{testPlayer2, testPlayer3})
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocX == 0)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocY == 0)

	// test that sending multiple moves in one tick works correctly
	moveMultiple1 := types.Move{"b", true, false, true, false, 2, 0.2} // up, down, left, right
	moveMultiple2 := types.Move{"b", false, true, false, true, 3, 0.2} // up, down, left, right
	game2.HandleMakeMove(moveMultiple1)
	game2.HandleMakeMove(moveMultiple2)
	game2.TickTock()

	testMove([]types.ModPlayer{testPlayer2, testPlayer3})
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-2].LocX == m[testPlayer2][len(m[testPlayer2])-1].LocX)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-2].LocY == m[testPlayer2][len(m[testPlayer2])-1].LocY)

	// test nearest player: choose a random configuration of players and verify that the players' healths are as expected
	// insert ATTACKPLAYERS players with non-dud weapons, tick, compare with expected player healths, and repeat until at most one player is left standing
	game2.HandlePlayerPop(testPlayer2)
	game2.HandlePlayerPop(testPlayer3)

	livePlayers := make(map[string]types.Triple[float64, float64, int])

	for i := 0; i < ATTACKPLAYERS; i++ {
		livePlayers[strconv.Itoa(i)] = types.Triple[float64, float64, int]{rand.Float64() * LENGTH, rand.Float64() * LENGTH, 100}
		game2.PushPlayer(component.PlayerComponent{strconv.Itoa(i), livePlayers[strconv.Itoa(i)].Third, 0, game2.Melee, types.Pair[float64, float64]{livePlayers[strconv.Itoa(i)].First, livePlayers[strconv.Itoa(i)].Second}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, types.Pair[float64, float64]{0, 0}, true, -1})
	}

	sim := func() bool {
		attacks := 0
		kill := make([]string, 0)
		for player, loc := range livePlayers {
			closestPlayer := "-1"
			for otherplayer, otherloc := range livePlayers {
				if otherplayer != player && (closestPlayer == "-1" || main.distance(loc, otherloc) < main.distance(loc, livePlayers[closestPlayer])) {
					closestPlayer = otherplayer
				}
			}

			if closestPlayer != "-1" && main.distance(loc, livePlayers[closestPlayer]) <= game2.Weapons[game2.Melee].Range {
				livePlayers[closestPlayer] = types.Triple[float64, float64, int]{livePlayers[closestPlayer].First, livePlayers[closestPlayer].Second, livePlayers[closestPlayer].Third - game2.Weapons[game2.Melee].Attack}
				attacks++
			}
		}

		for player, loc := range livePlayers {
			if loc.Third <= 0 {
				kill = append(kill, player)
			}
		}

		for _, killed := range kill {
			delete(livePlayers, killed)
		}

		return attacks == 0
	}

	comp := func() bool { // compares simulated player map to serverside player map; if
		serverMap := make(map[string]types.Triple[float64, float64, int])
		for i := 0; i < ATTACKPLAYERS; i++ {
			if p, err = game2.GetPlayerState(types.ModPlayer{strconv.Itoa(i)}); err == nil {
				serverMap[strconv.Itoa(i)] = types.Triple[float64, float64, int]{p.Loc.First, p.Loc.Second, p.Health}
			}
		}

		eq := len(serverMap) == len(livePlayers)
		for player, loc := range livePlayers {
			eq = eq && loc == serverMap[player]
		}

		return eq
	}

	for len(livePlayers) > 1 {
		if noMoreAttacks := sim(); noMoreAttacks {
			break
		}
		game2.TickTock()
		assert.Assert(t, comp())
		fmt.Println("Still attacking")
	}

	fmt.Println("Tests successfully passed")
}
