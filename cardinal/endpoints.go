package main

import (
	"github.com/argus-labs/new-game/components"
	"math"
	"net/http"

	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/new-game/utils"
)

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
