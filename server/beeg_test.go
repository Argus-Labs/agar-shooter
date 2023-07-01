package main

import (
	//"errors"
	"testing"
	"net/http"
	"fmt"

	//"github.com/argus-labs/world-engine/cardinal/ecs"
	"gotest.tools/v3/assert"
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
	game := Game{Pair[float64,float64]{LENGTH,LENGTH}, 2, []string{"a","b"}}
	var err error
	var player ModPlayer
	var contains bool
	var p PlayerComponent

	err = CreateGame(game)

	if err != nil {
		fmt.Println("pewp")
		fmt.Println("%w", err)
	}

	// test adding player moves and making player move each tick
	testPlayer1, testPlayer2, testPlayer3 := ModPlayer{"a"}, ModPlayer{"b"}, ModPlayer{"c"}

	m := make(map[ModPlayer] []BarePlayer)

	testMove := func(players []ModPlayer, a bool){
		for _, player = range players {
			_, contains = m[player]

			if !contains {
				m[player] = make([]BarePlayer, 0)
			}

			p, err = GetPlayerState(player)
			fmt.Println(player, ": ", p)
			assert.NilError(t, err)
			m[player] = append(m[player], p.Simplify())
		}
	}

	testMove([]ModPlayer{testPlayer1, testPlayer2}, false)

	move := Move{"a", true, false, true, false, 0, 0}// up, down, left, right
	HandleMakeMove(move)

	testMove([]ModPlayer{testPlayer1, testPlayer2}, false)

	TickTock()

	testMove([]ModPlayer{testPlayer1, testPlayer2}, false)

	assert.DeepEqual(t, m[testPlayer1][0], m[testPlayer1][1])
	assert.DeepEqual(t, m[testPlayer2][0], m[testPlayer2][1])

	assert.Assert(t, m[testPlayer1][1] != m[testPlayer1][2])
	assert.DeepEqual(t, m[testPlayer2][1], m[testPlayer2][2])

	// test adding a player to the game
	_, err = GetPlayerState(testPlayer3)
	assert.Assert(t, err != nil)

	fmt.Println("start push")
	HandlePlayerPush(testPlayer3)
	fmt.Println("end push")

	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3}, true)

	// test moving players after a new player has been added
	TickTock()
	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3}, false)

	assert.Assert(t, m[testPlayer1][len(m[testPlayer1]) - 2] != m[testPlayer1][len(m[testPlayer1]) - 1])
	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	assert.DeepEqual(t, m[testPlayer3][len(m[testPlayer3]) - 2], m[testPlayer3][len(m[testPlayer3]) - 1])

	newMove := Move{"c", false, true, false, true, 0, 0}
	HandleMakeMove(newMove)
	TickTock()

	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3}, false)

	assert.Assert(t, m[testPlayer1][len(m[testPlayer1]) - 2] != m[testPlayer1][len(m[testPlayer1]) - 1])
	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3]) - 2] != m[testPlayer3][len(m[testPlayer3]) - 1])

	// test moving players after a player has been removed
	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3}, true)

	fmt.Println("start pop")
	HandlePlayerPop(testPlayer1)
	fmt.Println("end pop")

	_, err = GetPlayerState(testPlayer1)
	assert.Assert(t, err != nil)

	TickTock()

	testMove([]ModPlayer{testPlayer2, testPlayer3}, false)

	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3]) - 2] != m[testPlayer3][len(m[testPlayer3]) - 1])

	//// test that players do not go beyond boundaries
	//p2X := m[testPlayer2][len(m[testPlayer2])-1].LocX
	//p3Y := m[testPlayer3][len(m[testPlayer3])-1].LocY

	//move1 := Move{"b", true, false, false, false, 0, 0}// up, down, left, right
	//move2 := Move{"c", false, false, false, true, 0, 0}

	//HandleMakeMove(move1)
	//HandleMakeMove(move2)
	//for i := 0; i < 5*LENGTH; i++ {
	//	TickTock()
	//}

	//testMove([]ModPlayer{testPlayer2, testPlayer3}, false)
	//assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocX == p2X)
	//assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-1].LocY == p3Y)
	//assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocY == LENGTH)
	//assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-1].LocX == LENGTH)

	//move3 := Move{"b", false, true, true, false, 0, 0}// up, down, left, right
	//HandleMakeMove(move3)

	//for i := 0; i < 5*LENGTH; i++ {
	//	TickTock()
	//}

	//testMove([]ModPlayer{testPlayer2, testPlayer3}, false)
	//assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocX == 0)
	//assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocY == 0)

	// test that sending multiple moves in one tick works correctly
	moveMultiple1 := Move{"b", true, false, true, false, 0, 0.4}// up, down, left, right
	moveMultiple2 := Move{"b", false, true, false, true, 0, 0.4}// up, down, left, right
	HandleMakeMove(moveMultiple1)
	HandleMakeMove(moveMultiple2)
	TickTock()
	
	testMove([]ModPlayer{testPlayer2, testPlayer3}, false)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-2].LocX == m[testPlayer2][len(m[testPlayer2])-1].LocX)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-2].LocY == m[testPlayer2][len(m[testPlayer2])-1].LocY)

	fmt.Println("Tests successfully passed")
}
