package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	OK                  = 0
	CANCELED            = 1
	UNKNOWN             = 2
	INVALID_ARGUMENT    = 3
	DEADLINE_EXCEEDED   = 4
	NOT_FOUND           = 5
	ALREADY_EXISTS      = 6
	PERMISSION_DENIED   = 7
	RESOURCE_EXHAUSTED  = 8
	FAILED_PRECONDITION = 9
	ABORTED             = 10
	OUT_OF_RANGE        = 11
	UNIMPLEMENTED       = 12
	INTERNAL            = 13
	UNAVAILABLE         = 14
	DATA_LOSS           = 15
	UNAUTHENTICATED     = 16
)

const (
	EnvGameServer = "GAME_SERVER_ADDR"
)

type MatchState struct {
	presences map[string]runtime.Presence// contains all in-game players
}

type Match struct{}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {// called when connection is established

	// Register the RPC function of Cardinal to Nakama to create a proxy
	err := InitializeCardinalProxy(logger, initializer)
	if err != nil {
		return err
	}

	// Create the singleton match
	err2 := initializer.RegisterMatch("singleton_match", newMatch)
	if err2 != nil {
		return err2
	}

	return nil
}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		presences: make(map[string]runtime.Presence),
	}

	tickRate := 5
	label := ""

	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {

	acceptUser := true

	return state, acceptUser, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		mState.presences[p.GetUserId()] = p
	}

	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		delete(mState.presences, p.GetUserId())
	}

	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	//broadcast; later, handle transactions here instead of in cardinal
	
	return mState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {

	return state
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}

func InitializeCardinalProxy(logger runtime.Logger, initializer runtime.Initializer) error {
	gameServerAddr := os.Getenv(EnvGameServer)
	if gameServerAddr == "" {
		msg := fmt.Sprintf("Must specify a game server via %s", EnvGameServer)
		logger.Error(msg)
		return errors.New(msg)
	}

	makeURL := func(resource string) string {
		return fmt.Sprintf("%s/%s", gameServerAddr, resource)
	}

	// Get the list of available endpoints from the backend server
	resp, err := http.Get(makeURL("list"))
	if err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	var endpoints []string
	if err := dec.Decode(&endpoints); err != nil {
		return err
	}

	for _, e := range endpoints {
		logger.Debug("registering: %v", e)
		currEndpoint := e
		err := initializer.RegisterRpc(e, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
			logger.Debug("Got request for %q", currEndpoint)

			req, err := http.NewRequestWithContext(ctx, "GET", makeURL(currEndpoint), strings.NewReader(payload))
			if err != nil {
				logger.Error("request setup failed for endpoint %q: %v", currEndpoint, err)
				return "", runtime.NewError("request setup failed", INTERNAL)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logger.Error("request failed for endpoint %q: %v", currEndpoint, err)
				return "", runtime.NewError("request failed", INTERNAL)
			}
			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				logger.Error("bad status code: %v: %s", resp.Status, body)
				return "", runtime.NewError("bad status code", INTERNAL)
			}
			str, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("can't read body")
				return "", runtime.NewError("read body failed", INTERNAL)
			}
			return string(str), nil
		})
		if err != nil {
			logger.Error("failed to register endpoint %q: %v", e, err)
		}
	}
	return nil
}
