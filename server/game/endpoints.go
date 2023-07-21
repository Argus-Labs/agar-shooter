package game

import (
	"fmt"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/component"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/types"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/utils"
	"math"
	"net/http"
)

func HandlePlayerPush(w http.ResponseWriter, r *http.Request) {
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

	err := _handlePlayerPush(player)

	if err != nil {
		utils.WriteError(w, "error pushing: ", err)
		return
	}

	utils.WriteResult(w, "Player registration successful")
}

func HandlePlayerPop(w http.ResponseWriter, r *http.Request) {
	player := types.ModPlayer{}

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name format given", err)
		return
	}

	if _, contains := Players[player.Name]; !contains {
		utils.WriteError(w, "player name does not exist", nil)
		return
	}

	_handlePlayerPop(player)

	utils.WriteResult(w, "Player removal successful")

}

func HandleMakeMove(w http.ResponseWriter, r *http.Request) {
	moves := types.Move{}

	if err := utils.Decode(r, &moves); err != nil {
		utils.WriteError(w, "invalid move or player name given", err)
		return
	}

	_handleMakeMove(moves)

	utils.WriteResult(w, "move registered")
}

func GetPlayerState(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	comp, err := _getPlayerState(player)
	bareplayer := comp.Simplify()

	if err != nil {
		utils.WriteError(w, "could not get player state", err)
		return
	}

	utils.WriteResult(w, bareplayer)
}

func GetPlayerCoins(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	coins := NearbyCoins(player)

	utils.WriteResult(w, coins)
}

func GetPlayerStatus(w http.ResponseWriter, r *http.Request) { // get all locations of players --- array of pairs of strings and location (coordinate pairs)
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	comp, err := _getPlayerState(player)

	if err != nil {
		utils.WriteError(w, "could not get player state", err)
		return
	}

	utils.WriteResult(w, comp)
}

func CheckExtraction(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	coins := _checkExtraction(player)

	utils.WriteResult(w, coins)
}

func GetExtractionPoint(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	extract := _getExtractionPoint(player)

	utils.WriteResult(w, fmt.Sprintf("{\"X\":%f, \"Y\":%f}", extract.First, extract.Second))
}

func TestAddHealth(w http.ResponseWriter, r *http.Request) {
	var player types.ModPlayer

	if err := utils.Decode(r, &player); err != nil {
		utils.WriteError(w, "invalid player name given", err)
		return
	}

	PlayerComp.Update(World, Players[player.Name], func(comp component.PlayerComponent) component.PlayerComponent { // modifies player location
		comp.Health = int(math.Max(float64(comp.Health+10), 100.))

		return comp
	})

	utils.WriteResult(w, "health added")
}

func RecentAttacks(w http.ResponseWriter, r *http.Request) {
	attacks := _recentAttacks()
	utils.WriteResult(w, attacks)
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	_game := types.Game{types.Pair[float64, float64]{100, 100}, 2, []string{}} //"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}}// removed {"a","b"}
	if err := _createGame(_game); err != nil {
		utils.WriteError(w, "error initializing game", err)
	}

	for i := 0; i < InitRepeatSpawn; i++ {
		go SpawnCoins()
	}

	utils.WriteResult(w, "game created")
}

func Tig(w http.ResponseWriter, r *http.Request) {
	if err := TickTock(); err != nil {
		utils.WriteError(w, "error ticking", err)
	}

	if err := SpawnCoins(); err != nil {
		utils.WriteError(w, "error spawning coins", err)
	}

	utils.WriteResult(w, "game tick completed; coins spawned")
}
