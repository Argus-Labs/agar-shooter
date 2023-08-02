package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"fmt"
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
		{"games/push", handlePlayerPush},
		{"games/pop", handlePlayerPop},
		{"games/move", handleMakeMove},
		{"games/state", getPlayerState},
		{"games/coins", getPlayerCoins},
		{"games/health", getPlayerHealth},
		{"games/tick", tig},
		{"games/create", createGame},
		{"games/totalcoins", getCoins},
		{"games/attacks", recentAttacks},
		{"games/testaddhealth", testAddHealth},
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
			writeError(w, "can't marshal list", err)
		}
	})

	log.Printf("Starting server on port %s\n", port)

	http.ListenAndServe(":"+port, nil)
}

func writeError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(500)// error
	fmt.Fprintf(w, "%s: %v", msg, err)
}

func writeResult(w http.ResponseWriter, v any) {
	w.WriteHeader(200)// success
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
