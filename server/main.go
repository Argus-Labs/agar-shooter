package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"fmt"
	"math"

	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

const EnvGameServerPort = "GAME_SERVER_PORT"// test

func main() {
	// opens port
	port := os.Getenv(EnvGameServerPort)
	if port == "" {
		fmt.Errorf("Must specify a port via %s", EnvGameServerPort)
	}



	// defines an array of handlers that do one of handle games, create games, and make moves
	handlers := []struct {
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"games/push", handlePlayerPush},
		{"games/pop", handlePlayerPop},
		{"games/move", handleMakeMove},
		{"games/status", getPlayerState},
		{"games/tick", tig},
		{"games/create", createGame},
	}

	log.Printf("Attempting to register %d handlers\n", len(handlers))
	// handles the function by taking the response, figuring out which game function to call, and calling it
	paths := []string{}


	for _, h := range handlers {
		http.HandleFunc("/"+h.path, h.handler)
		paths = append(paths, h.path)
	}
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)

		if err := enc.Encode(paths); err != nil {
			writeError(w, "can't marshal list", err)
		}

	})

	log.Printf("Starting server on port %s\n", port)

	http.ListenAndServe(":"+port, nil)
}

func writeError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(500)
	fmt.Fprintf(w, "%s: %v", msg, err)
}

func writeResult(w http.ResponseWriter, v any) {
	if s, ok := v.(string); ok {
		v = struct{ Msg string }{Msg: s}
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		writeError(w, "can't encode", err)
		return
	}
}

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

//// need to figure out how to use these three functions; just need to modify the functions/endpoints we use and the clientside nakama code handles the rest
//func handleCreateGame(w http.ResponseWriter, r *http.Request) {
//	var gameData Game
//	if err := decode(r, &gameData); err != nil {
//		writeError(w, "decode failed", err)
//		return
//	}
//
//	for _, playername := range GameParams.Players {
//		if playername == "" {
//			writeError(w, "must name all players", nil)
//			return
//		}
//	}
//
//	// only creates things after checking that there are no errors with the gameParams input
//	err := HandleCreateGame(gameData)
//
//	if err != nil {
//		writeError(w, "error initializing world: ", err)
//		return
//	}
//
//	w.WriteHeader(200)// outputs success
//	writeResult(w, "success")// also write the location of each player by playername
//}

func handlePlayerPush(w http.ResponseWriter, r *http.Request) {// adds player to world
	player := ModPlayer{}

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name format given: ", err)
		return
	}

	if _, contains := Players[player.Name]; contains {
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

func handlePlayerPop(w http.ResponseWriter, r *http.Request) {// removes player from world
	player := ModPlayer{}

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name format given", err)
		return
	}

	if _, contains := Players[player.Name]; !contains {
		writeError(w, "player name does not exist", nil)
		return
	}

	writeResult(w, "Player removal successful")

}

// write player addition, removal, and update loop
func handleMakeMove(w http.ResponseWriter, r *http.Request) {// add move to transaction queue
	moves := Move{}

	if err := decode(r, &moves); err != nil {
		writeError(w, "invalid move or player name given", err)
		return
	}

	HandleMakeMove(moves)

	writeResult(w, "move registered")// also write the location of each player by playername
}

func getPlayerState(w http.ResponseWriter, r *http.Request) {// use in place of broadcast to get player state for now
	var player ModPlayer

	if err := decode(r, &player); err != nil {
		writeError(w, "invalid player name given", err)
		return
	}

	playercomp, err := GetPlayerState(player)

	if err != nil {
		writeError(w, "could not get player state", err)
		return
	}

	writeResult(w, playercomp)// convert to string
}

func createGame(w http.ResponseWriter, r *http.Request) {
	game := Game{Pair[float64,float64]{1000,1000}, 1, []string{"a", "b"}}
	errr := CreateGame(game)// move this to somewhere with an http.ResponseWriter
	if errr != nil {// error from game creation
		writeError(w, "error initializing game", errr)
	}
}

func CreateGame(game Game) error {
	//if World.stateIsLoaded {
	//	return fmt.Errorf("already loaded state")
	//}
	GameParams = game
	World.RegisterComponents(ItemMapComp, PlayerMapComp, PlayerComp)
	World.AddSystem(processMoves)
	World.AddSystem(makeMoves)

	World.LoadGameState()
	MoveTx.SetID(0)
	ItemMap, err := World.Create(ItemMapComp)// creates an ItemMap entity
	PlayerMap, err := World.Create(PlayerMapComp)// creates a PlayerMap entity
	playerIDs, err := World.CreateMany(len(GameParams.Players), PlayerComp)// creates player entities

	for i, playername := range GameParams.Players {// associates storage.EntityIDs with each player
		Players[playername] = playerIDs[i]
	}

	if err != nil {
		return fmt.Errorf("Error initializing game objects: %w", err)
	}

	// initializes player and item maps
	itemmap := make(map[Pair[int, int]] map[Pair[storage.EntityID, Pair[float64,float64]]] void)
	playermap := make(map[Pair[int, int]] map[Pair[storage.EntityID, Pair[float64, float64]]] void)
	for i := 0; i <= int(math.Ceil(GameParams.Dims.First/GameParams.CSize)); i++ {
		for j := 0; j <= int(math.Ceil(GameParams.Dims.Second/GameParams.CSize)); j++ {
			itemmap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64,float64]]] void)
			playermap[Pair[int,int]{i,j}] = make(map[Pair[storage.EntityID, Pair[float64, float64]]] void)
		}
	}

	ItemMapComp.Set(World, ItemMap, ItemMapComponent{itemmap})// initializes ItemMap using empty map
	PlayerMapComp.Set(World, ItemMap, PlayerMapComponent{playermap})// initializes PlayerMap using empty map

	for _, playername := range GameParams.Players {
		PlayerComp.Set(World, Players[playername], PlayerComponent{playername, 100, 0, Melee, Pair[float64,float64]{25,25}, Direction{90, Pair[float64,float64]{0,0}}})// initializes player entitities through their component

		PlayerMapComp.Update(World, PlayerMap, func(comp PlayerMapComponent) PlayerMapComponent {// adds players to the board
			playercomp, err := PlayerComp.Get(World, Players[playername])

			if err != nil {
				fmt.Errorf("Error getting location with callback function: %w", err)
				return comp
			}

			newPlayer := Pair[storage.EntityID, Pair[float64,float64]]{Players[playername], playercomp.Loc}
			comp.Players[Pair[int,int]{25/int(GameParams.CSize),25/int(GameParams.CSize)}][newPlayer] = pewp

			return comp
		})
	}

	//World.RegisterTransactions(MoveTx)

	// calls callback goroutine to keep World ticking

	//go func(){ TODO enable after testing
	//	for range time.Tick(time.Second/tickRate) {
	//		World.Tick()
	//	}
	//}()

	return nil
}

func tig(w http.ResponseWriter, r *http.Request) {TickTock();}// use in place of broadcast to tick
