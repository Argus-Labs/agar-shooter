package main

import (
	"net/http"
	"fmt"
	"math"
)

func handlePlayerPush(w http.ResponseWriter, r *http.Request) {
	player := AddPlayer{}

	fmt.Println(r)

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name format given: ", err)
		return
	}

	if _, contains := Players[player.Name]; contains {
		return
		writeError(w, "player name already exists", nil)
		return
	}

	err := HandlePlayerPush(player)

	if err != nil {
		writeError(w, "error pushing: ", err)
		return
	}

	writeResult(w, "Player registration successful")
}

func handlePlayerPop(w http.ResponseWriter, r *http.Request) {
	player := ModPlayer{}

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name format given", err)
		return
	}

	if _, contains := Players[player.Name]; !contains {
		writeError(w, "player name does not exist", nil)
		return
	}

	HandlePlayerPop(player)

	writeResult(w, "Player removal successful")

}

func handleMakeMove(w http.ResponseWriter, r *http.Request) {
	moves := Move{}

	if err := decode(r, &moves); err != nil {
		writeError(w, "invalid move or player name given", err)
		return
	}

	HandleMakeMove(moves)

	writeResult(w, "move registered")
}

func getPlayerState(w http.ResponseWriter, r *http.Request) {
	var player ModPlayer

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name given", err)
		return
	}

	comp, err := GetPlayerState(player)
	bareplayer := comp.Simplify()

	if err != nil {
		writeError(w, "could not get player state", err)
		return
	}

	writeResult(w, bareplayer)
}

func getPlayerCoins(w http.ResponseWriter, r *http.Request) {
	var player ModPlayer

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name given", err)
		return
	}

	coins := NearbyCoins(player)

	writeResult(w, coins)
}

func getPlayerStatus(w http.ResponseWriter, r *http.Request) {// get all locations of players --- array of pairs of strings and location (coordinate pairs)
	var player ModPlayer

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name given", err)
		return
	}

	comp, err := GetPlayerState(player)

	if err != nil {
		writeError(w, "could not get player state", err)
		return
	}

	writeResult(w, comp)
}

func checkExtraction(w http.ResponseWriter, r *http.Request) {
	var player ModPlayer

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name given", err)
		return
	}

	coins := CheckExtraction(player)

	writeResult(w, coins)
}

func testAddHealth(w http.ResponseWriter, r *http.Request) {
	var player ModPlayer

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name given", err)
		return
	}

	PlayerComp.Update(World, Players[player.Name], func(comp PlayerComponent) PlayerComponent{// modifies player location
		comp.Health = int(math.Max(float64(comp.Health+10), 100.))
		
		return comp
	})

	writeResult(w, "health added")
}

func recentAttacks(w http.ResponseWriter, r *http.Request) {
	attacks := RecentAttacks()
	writeResult(w, attacks)
}

func createGame(w http.ResponseWriter, r *http.Request) {
	game := Game{Pair[float64,float64]{100,100}, 5, []string{}}// removed {"a","b"}
	if err := CreateGame(game); err != nil {
		writeError(w, "error initializing game", err)
	}

	for i := 0; i < 5; i++ {
		go SpawnCoins(globalMut)
	}

	writeResult(w, "game created")
}

func tig(w http.ResponseWriter, r *http.Request) {
	if err := TickTock(); err != nil {
		writeError(w, "error ticking", err)
	}

	if err := SpawnCoins(globalMut); err != nil {
		writeError(w, "error spawning coins", err)
	}

	writeResult(w, "game tick completed; coins spawned")
}