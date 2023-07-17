package main

import (
	"fmt"
	"math"
	"net/http"

	"github.com/argus-labs/new-game/components"
	msg "github.com/argus-labs/new-game/msg/tx"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/new-game/utils"
)

func handlePlayerPush(w http.ResponseWriter, r *http.Request) {
	player := types.AddPlayer{}

	fmt.Println(r)

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name format given: ", err)
		return
	}

	if _, contains := Players[player.Name]; contains {
		return
		utils.WriteError(w, "player name already exists", nil)
		return
	}

	err := HandlePlayerPush(player)

	if err != nil {
		utils.WriteError(w, "error pushing: ", err)
		return
	}

	utils.WriteResult(w, "Player registration successful")
}

func handlePlayerPop(w http.ResponseWriter, r *http.Request) {
	player := types.ModPlayer{}

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name format given", err)
		return
	}

	if _, contains := Players[player.Name]; !contains {
		utils.WriteError(w, "player name does not exist", nil)
		return
	}

	HandlePlayerPop(player)

	utils.WriteResult(w, "Player removal successful")

}

func handleMakeMove(w http.ResponseWriter, r *http.Request) {
	moves := msg.MovePlayerMsg{}

	if err := utils.Decode(r, &moves); err != nil {
		utils.WriteError(w, "invalid move or player name given", err)
		return
	}

	HandleMakeMove(moves)

	utils.WriteResult(w, "move registered")
}

func getPlayerState(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	comp, err := GetPlayerState(player)
	bareplayer := comp.Simplify()

	if err != nil {
		utils.WriteError(w, "could not get player state", err)
		return
	}

	utils.WriteResult(w, bareplayer)
}

func getPlayerCoins(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	coins := NearbyCoins(player)

	utils.WriteResult(w, coins)
}

func getPlayerStatus(w http.ResponseWriter, r *http.Request) { // get all locations of players --- array of types.Pairs of strings and location (coordinate types.Pairs)
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	comp, err := GetPlayerState(player)

	if err != nil {
		utils.WriteError(w, "could not get player state", err)
		return
	}

	utils.WriteResult(w, comp)
}

func checkExtraction(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	coins := CheckExtraction(player)

	utils.WriteResult(w, coins)
}

func testAddHealth(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	components.Player.Update(World, Players[player.Name], func(comp components.PlayerComponent) components.PlayerComponent { // modifies player location
		comp.Health = int(math.Max(float64(comp.Health+10), 100.))

		return comp
	})

	utils.WriteResult(w, "health added")
}

func recentAttacks(w http.ResponseWriter, r *http.Request) {
	attacks := RecentAttacks()
	utils.WriteResult(w, attacks)
}

func createGame(w http.ResponseWriter, r *http.Request) {
	game := Game{types.Pair[float64, float64]{100, 100}, 5, []string{}} // removed {"a","b"}
	if err := CreateGame(game); err != nil {
		utils.WriteError(w, "error initializing game", err)
	}

	for i := 0; i < InitRepeatSpawn; i++ {
		go SpawnCoins()
	}

	utils.WriteResult(w, "game created")
}

func tig(w http.ResponseWriter, r *http.Request) {
	if err := TickTock(); err != nil {
		utils.WriteError(w, "error ticking", err)
	}

	if err := SpawnCoins(); err != nil {
		utils.WriteError(w, "error spawning coins", err)
	}

	utils.WriteResult(w, "game tick completed; coins spawned")
}
