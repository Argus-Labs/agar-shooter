package tests

import (
	//"errors"
	"testing"
	"fmt"

	//"github.com/argus-labs/world-engine/cardinal/ecs"
	"gotest.tools/v3/assert"
	. "github.com/argus-labs/world-engine/game/sample_game_server/server"
)

func TestPushPop(t *testing.T) {
	gameParams := Game{Pair[int,int]{1000,1000}, 2, []string{"d","e"}}
	var err error
	var player ModPlayer
	var contains bool
	var p PlayerComponent

	HandleCreateGame(gameParams)

	// test board correctly initialized
	testPlayer1, testPlayer2, testPlayer3 := ModPlayer{"d"}, ModPlayer{"e"}, ModPlayer{"f"}

	m := make(map[ModPlayer] []PlayerComponent)

	testPlay := func(players []ModPlayer){
		for _, player = range players {
			_, contains = m[player]
			
			if !contains {
				m[player] = make([]PlayerComponent, 0)
			}

			p, err = GetPlayerState(player)
			fmt.Println(player, ": ", p)
			assert.NilError(t, err)
			m[player] = append(m[player], p)
		}
	}

	testPlay([]ModPlayer{testPlayer1, testPlayer2})
	testPlay([]ModPlayer{testPlayer1, testPlayer2})
	assert.DeepEqual(t, m[testPlayer1][len(m[testPlayer1]) - 2], m[testPlayer1][len(m[testPlayer1]) - 1])
	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])

	// test adding a player to the game
	_, err = GetPlayerState(testPlayer3)
	assert.Assert(t, err != nil)

	fmt.Println("start push")
	HandlePlayerPush(testPlayer3)
	fmt.Println("end push")

	testPlay([]ModPlayer{testPlayer1, testPlayer2, testPlayer3})
	//assert.DeepEqual(t, m[testPlayer1][len(m[testPlayer1]) - 2], m[testPlayer1][len(m[testPlayer1]) - 1])
	//assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])

	testPlay([]ModPlayer{testPlayer1, testPlayer2, testPlayer3})
	//assert.DeepEqual(t, m[testPlayer1][len(m[testPlayer1]) - 2], m[testPlayer1][len(m[testPlayer1]) - 1])
	//assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	//assert.DeepEqual(t, m[testPlayer3][len(m[testPlayer3]) - 2], m[testPlayer3][len(m[testPlayer3]) - 1])
	
	// test moving players after a player has been removed
	fmt.Println("start pop")
	HandlePlayerPop(testPlayer1)
	fmt.Println("end pop")

	_, err = GetPlayerState(testPlayer1)
	assert.Assert(t, err != nil)

	TickTock()

	testPlay([]ModPlayer{testPlayer2, testPlayer3})
	//assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	//assert.DeepEqual(t, m[testPlayer3][len(m[testPlayer3]) - 2], m[testPlayer3][len(m[testPlayer3]) - 1])

	fmt.Println("Tests successfully passed")
}
