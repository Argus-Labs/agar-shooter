package main

import (
	"encoding/json"
	"fmt"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/game"
	"github.com/argus-labs/world-engine/game/sample_game_server/server/utils"
	"log"
	"net/http"
	"os"
)

const EnvGameServerPort = "GAME_SERVER_PORT"

func main() {
	port := os.Getenv(EnvGameServerPort)
	if port == "" {
		fmt.Errorf("Must specify a port via %s", EnvGameServerPort)
	}

	handlers := []struct {
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"games/push", game.HandlePlayerPush},
		{"games/pop", game.HandlePlayerPop},
		{"games/move", game.HandleMakeMove},
		{"games/state", game.GetPlayerState},
		{"games/status", game.GetPlayerStatus},
		{"games/coins", game.GetPlayerCoins},
		{"games/tick", game.Tig},
		{"games/create", game.CreateGame},
		{"games/offload", game.CheckExtraction},
		{"games/extract", game.GetExtractionPoint},
		{"games/attacks", game.RecentAttacks},
		{"games/testaddhealth", game.TestAddHealth},
	}

	log.Printf("Attempting to register %d handlers\n", len(handlers))
	paths := []string{}

	for _, h := range handlers {
		http.HandleFunc("/"+h.path, h.handler)
		paths = append(paths, h.path)
	}
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)

		if err := enc.Encode(paths); err != nil {
			utils.WriteError(w, "can't marshal list", err)
		}
	})

	log.Printf("Starting server on port %s\n", port)

	http.ListenAndServe(":"+port, nil)
}
