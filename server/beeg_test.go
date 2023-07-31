package main

import (
	"testing"
	"net/http"
	"fmt"
	"math/rand"
	"strconv"
	"time"

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
	const ATTACKPLAYERS = 100
	const ASYNCATTACKWAIT = 5
	game := Game{Pair[float64,float64]{LENGTH,LENGTH}, 1, []string{}}
	
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
	weapon1, _ := World.Create(WeaponComp)
	WeaponComp.Set(World, weapon1, WeaponComponent{Pair[float64, float64]{-1,-1}, Dud, Weapons[Dud].MaxAmmo, 0})
	PushPlayer(PlayerComponent{"a", 100, 0, weapon1, Pair[float64,float64]{250,250}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{500, 500}, time.Now().UnixNano(), true, -1, 0})

	weapon2, _ := World.Create(WeaponComp)
	WeaponComp.Set(World, weapon2, WeaponComponent{Pair[float64, float64]{-1,-1}, Dud, Weapons[Dud].MaxAmmo, 0})
	PushPlayer(PlayerComponent{"b", 100, 0, weapon2, Pair[float64,float64]{750,750}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{250, 250}, time.Now().UnixNano(), true, -1, 0})

	m := make(map[ModPlayer] []TestPlayer)

	testMove := func(players []ModPlayer){
		for _, player = range players {
			_, contains = m[player]

			if !contains {
				m[player] = make([]TestPlayer, 0)
			}

			p, err = GetPlayerState(player)
			fmt.Println(player, ": ", p)
			assert.NilError(t, err)
			m[player] = append(m[player], p.Testify())
		}
	}

	testMove([]ModPlayer{testPlayer1, testPlayer2})

	move := Move{"a", true, false, true, false, 0, 0.2}// up, down, left, right
	HandleMakeMove(move)

	testMove([]ModPlayer{testPlayer1, testPlayer2})

	TickTock()

	testMove([]ModPlayer{testPlayer1, testPlayer2})

	assert.DeepEqual(t, m[testPlayer1][0], m[testPlayer1][1])
	assert.DeepEqual(t, m[testPlayer2][0], m[testPlayer2][1])

	assert.Assert(t, m[testPlayer1][1] != m[testPlayer1][2])
	assert.DeepEqual(t, m[testPlayer2][1], m[testPlayer2][2])

	// test adding a player to the game
	_, err = GetPlayerState(testPlayer3)
	assert.Assert(t, err != nil)

	fmt.Println("start push")
	weapon3, _ := World.Create(WeaponComp)
	WeaponComp.Set(World, weapon3, WeaponComponent{Pair[float64, float64]{-1,-1}, Dud, Weapons[Dud].MaxAmmo, 0})
	PushPlayer(PlayerComponent{"c", 100, 0, weapon3, Pair[float64,float64]{500,500}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{750, 750}, time.Now().UnixNano(), true, -1, 0})
	fmt.Println("end push")

	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	// test moving players after a new player has been added
	TickTock()
	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	fmt.Println(m[testPlayer1])
	assert.Assert(t, m[testPlayer1][len(m[testPlayer1]) - 2] != m[testPlayer1][len(m[testPlayer1]) - 1])
	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	assert.DeepEqual(t, m[testPlayer3][len(m[testPlayer3]) - 2], m[testPlayer3][len(m[testPlayer3]) - 1])

	newMove := Move{"c", false, true, false, true, 0, 0.2}
	HandleMakeMove(newMove)
	TickTock()

	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	assert.Assert(t, m[testPlayer1][len(m[testPlayer1]) - 2] != m[testPlayer1][len(m[testPlayer1]) - 1])
	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3]) - 2] != m[testPlayer3][len(m[testPlayer3]) - 1])

	// test moving players after a player has been removed
	testMove([]ModPlayer{testPlayer1, testPlayer2, testPlayer3})

	fmt.Println("start pop")
	HandlePlayerPop(testPlayer1)
	fmt.Println("end pop")

	_, err = GetPlayerState(testPlayer1)
	assert.Assert(t, err != nil)

	TickTock()

	testMove([]ModPlayer{testPlayer2, testPlayer3})

	assert.DeepEqual(t, m[testPlayer2][len(m[testPlayer2]) - 2], m[testPlayer2][len(m[testPlayer2]) - 1])
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3]) - 2] != m[testPlayer3][len(m[testPlayer3]) - 1])

	// test that players do not go beyond boundaries
	p2X := m[testPlayer2][len(m[testPlayer2])-1].LocX
	p3Y := m[testPlayer3][len(m[testPlayer3])-1].LocY

	move1 := Move{"b", true, false, false, false, 0, 0.2}// up, down, left, right
	move2 := Move{"c", false, false, false, true, 1, 0.2}

	HandleMakeMove(move1)
	HandleMakeMove(move2)
	for i := 0; i < 5*LENGTH; i++ {
		TickTock()
	}

	testMove([]ModPlayer{testPlayer2, testPlayer3})
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocX == p2X)
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-1].LocY == p3Y)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocY == LENGTH)
	assert.Assert(t, m[testPlayer3][len(m[testPlayer3])-1].LocX == LENGTH)

	move3 := Move{"b", false, true, true, false, 1, 0.2}// up, down, left, right
	HandleMakeMove(move3)

	for i := 0; i < 5*LENGTH; i++ {
		TickTock()
	}

	testMove([]ModPlayer{testPlayer2, testPlayer3})
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocX == 0)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-1].LocY == 0)

	// test that sending multiple moves in one tick works correctly
	moveMultiple1 := Move{"b", true, false, true, false, 2, 0.2}// up, down, left, right
	moveMultiple2 := Move{"b", false, true, false, true, 3, 0.2}// up, down, left, right
	HandleMakeMove(moveMultiple1)
	HandleMakeMove(moveMultiple2)
	TickTock()
	
	testMove([]ModPlayer{testPlayer2, testPlayer3})
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-2].LocX == m[testPlayer2][len(m[testPlayer2])-1].LocX)
	assert.Assert(t, m[testPlayer2][len(m[testPlayer2])-2].LocY == m[testPlayer2][len(m[testPlayer2])-1].LocY)

	// test nearest player: choose a random configuration of players and verify that the players' healths are as expected
	// insert ATTACKPLAYERS players with non-dud weapons, tick, compare with expected player healths, and repeat until at most one player is left standing
	HandlePlayerPop(testPlayer2)
	HandlePlayerPop(testPlayer3)

	livePlayers := make(map[string] Triple[float64, float64, int])

	for i := 0; i < ATTACKPLAYERS; i++ {
		livePlayers[strconv.Itoa(i)] = Triple[float64, float64, int]{rand.Float64()*LENGTH, rand.Float64()*LENGTH, 100}
		weaponi, _ := World.Create(WeaponComp)
		WeaponComp.Set(World, weaponi, WeaponComponent{Pair[float64, float64]{-1,-1}, TestWeapon, Weapons[TestWeapon].MaxAmmo, 0})
		PushPlayer(PlayerComponent{strconv.Itoa(i), livePlayers[strconv.Itoa(i)].Third, 0, weaponi, Pair[float64, float64]{livePlayers[strconv.Itoa(i)].First, livePlayers[strconv.Itoa(i)].Second}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0, 0}, time.Now().UnixNano(), true, -1, 0})
	}

	sim := func() bool {
		attacks := 0
		kill := make([]string, 0)
		for player, loc := range livePlayers {
			closestPlayer := "-1"
			for otherplayer, otherloc := range livePlayers {
				if otherplayer != player && (closestPlayer == "-1" || distance(loc, otherloc) < distance(loc, livePlayers[closestPlayer])) {
					closestPlayer = otherplayer
				}
			}

			if closestPlayer != "-1" && distance(loc, livePlayers[closestPlayer]) <= Weapons[TestWeapon].Range {
				livePlayers[closestPlayer] = Triple[float64, float64, int]{livePlayers[closestPlayer].First, livePlayers[closestPlayer].Second, livePlayers[closestPlayer].Third - Weapons[TestWeapon].Attack}
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

	comp := func() bool {// compares simulated player map to serverside player map; if 
		serverMap := make(map[string] Triple[float64, float64, int])
		for i := 0; i < ATTACKPLAYERS; i++ {
			if p, err = GetPlayerState(ModPlayer{strconv.Itoa(i)}); err == nil {
				serverMap[strconv.Itoa(i)] = Triple[float64, float64, int]{p.Loc.First, p.Loc.Second, p.Health}
			}
		}

		eq := len(serverMap) == len(livePlayers)
		for player, loc := range livePlayers {
			eq = eq && loc == serverMap[player]
		}
		
		return eq
	}

	for len(livePlayers) > 1 {
		TickTock()
		noMoreAttacks := sim()
		assert.Assert(t, comp())
		fmt.Println("Still attacking")
		if noMoreAttacks {
			break
		}
	}

	for player, _ := range livePlayers {
		HandlePlayerPop(ModPlayer{player})
	}

	// check that weapon ammo works as intended, check that reload times work
	// spawn players within some radius of each other, make them attack each other with one given an unlimited ammo weapon and the other given a limited ammo weapon, then check that each drops each other's health as expected and that the number of attacks is as expected

	weapon1, _ = World.Create(WeaponComp)
	WeaponComp.Set(World, weapon1, WeaponComponent{Pair[float64, float64]{-1,-1}, Melee, Weapons[Melee].MaxAmmo, 0})
	PushPlayer(PlayerComponent{"a", 100, 0, weapon1, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{500, 500}, time.Now().UnixNano(), true, -1, 0})

	weapon2, _ = World.Create(WeaponComp)
	WeaponComp.Set(World, weapon2, WeaponComponent{Pair[float64, float64]{-1,-1}, Melee, Weapons[Melee].MaxAmmo, 0})
	PushPlayer(PlayerComponent{"b", 100, 0, weapon2, Pair[float64,float64]{1,1}, Pair[float64,float64]{0,0}, Pair[float64,float64]{0,0}, Pair[float64,float64]{250, 250}, time.Now().UnixNano(), true, -1, 0})
	TickTock()
	wipun2, _ := WeaponComp.Get(World, weapon2)
	startTime := wipun2.LastAttack

	p, err = GetPlayerState(testPlayer1)
	p1Health := p.Health

	p, err = GetPlayerState(testPlayer1)
	p2Health := p.Health

	fmt.Println("attacks\n\n\n\n\n\n\n ")
	for i := time.Now().UnixNano(); i < startTime + 5*time.Second.Nanoseconds(); i = time.Now().UnixNano(){
		TickTock()
		p, err = GetPlayerState(testPlayer1)
		assert.Assert(t, p1Health - int(time.Duration(i - startTime).Seconds())*Weapons[Melee].Attack - p.Health <= Weapons[Melee].Attack)
		
		p, err = GetPlayerState(testPlayer2)
		p, err = GetPlayerState(testPlayer1)
		assert.Assert(t, p2Health - int(time.Duration(i - startTime).Seconds())*Weapons[Melee].Attack - p.Health <= Weapons[Melee].Attack)
		

		wipun1, _ := WeaponComp.Get(World, weapon1)
		wipun2, _ := WeaponComp.Get(World, weapon2)
		fmt.Println("wipun1:", wipun1)
		fmt.Println("wipun2:", wipun2)
		fmt.Println("TICK COMPLETED", time.Now().Unix())
	}

	fmt.Println("Tests successfully passed")
}
