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
	OK int64					= 0
	CANCELED int64				= 1
	UNKNOWN int64				= 2
	INVALID_ARGUMENT int64		= 3
	DEADLINE_EXCEEDED int64		= 4
	NOT_FOUND int64				= 5
	ALREADY_EXISTS int64		= 6
	PERMISSION_DENIED int64		= 7
	RESOURCE_EXHAUSTED int64	= 8
	FAILED_PRECONDITION int64	= 9
	ABORTED int64				= 10
	OUT_OF_RANGE int64			= 11
	UNIMPLEMENTED int64			= 12
	INTERNAL int				= 13
	UNAVAILABLE int64			= 14
	DATA_LOSS int64				= 15
	UNAUTHENTICATED int64		= 16
	MOVE int64					= 17
)

const (
	EnvGameServer = "GAME_SERVER_ADDR"
)

var (
	CallRPCs = make(map[string] func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error))// contains all RPC endpoint functions
)

type MatchState struct {// contains match data as the match progresses
	presences map[string]runtime.Presence// contains all in-game players
}

type Match struct{
	tick int
}// should contain data on the match, but because this is being handled through Cardinal, it's an empty struct that does nothing --- all match functions call Cardinal endpoints rather than actually updating some Nakama match state

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {// called when connection is established
	if err := InitializeCardinalProxy(logger, initializer); err != nil {// Register the RPC function of Cardinal to Nakama to create a proxy
		return err
	}

	// Create the singleton match
	if err := initializer.RegisterMatch("singleton_match", newMatch); err != nil {
		return err
	}

	matchId, err := nk.MatchCreate(ctx, "singleton_match", map[string]interface{}{})// calls the newMatch function, then calls MatchInit on the result

	if err != nil {
		return err
	}

	logger.Debug("MATCH ID: ", matchId)


	//if _, err := newMatch(ctx, logger, db, nk); err != nil {// stack overflow occurs here
	//	logger.Error(fmt.Errorf("Nakama: peepunable to create match; ", err).Error())
	//}

	//makeMatch := func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
	//	matchId, err := nk.MatchCreate(ctx, "lobby", map[string]interface{}{"invited": entries})

	//	if err != nil {
	//		for i := 0; i < 20; i++ {logger.Debug(">>>>pewp")}
	//		logger.Error(fmt.Errorf("Nakama: unable to create match; ", err).Error())
	//		return "", fmt.Errorf("Nakama: unable to create match; ", err)
	//	}
	//
	//	for i := 0; i < 20; i++ {logger.Debug("<<<<")}
	//	logger.Debug("MATCH ID: ", matchId)

	//	return matchId, nil
	//}

	//if err := initializer.RegisterMatchmakerMatched(makeMatch); err != nil {
	//	logger.Error("unable to register matchmaker matched hook: %v", err)
	//	return err
	//}

	//matchId, err := makeMatch(ctx, logger, db, nk, []runtime.MatchmakerEntry{})
	//logger.Debug("MATCH ID: ", matchId)
	//if err != nil {
	//	for i := 0; i < 20; i++ {logger.Debug(">>>>uhoh")}
	//	logger.Error(fmt.Errorf("Nakama: unable to create match; ", err).Error())
	//	return fmt.Errorf("Nakama: unable to create match; ", err)
	//}

	return nil
}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	//for i := 0; i < 20; i++ {logger.Debug(">>>>")}
	//matchId, err := nk.MatchCreate(ctx, "singleton_match", map[string]interface{}{})// error occurs here; should call MatchInit, but it doesn't because the debug messages don't print, so there's an error with actually calling this function
	//// the error message occurs because the nakama RuntimeModule's matchCreateFn outputs empty on the input argument
	//// to register the function, we call the NewRuntimeProviderGo function, which sets the matchCreateFn to matchprovider.CreateMatch. This function takes as input a RuntimeMatchCreateFunction that runs on the inputs to the CreateMatch function.

	//if err != nil {
	//	for i := 0; i < 20; i++ {logger.Debug("vvvv")}
	//	return &Match{0}, fmt.Errorf("Nakama: pewpunable to create match; ", err)// error returns here
	//}

	//for i := 0; i < 20; i++ {logger.Debug("<<<<")}
	//logger.Debug("MATCH ID: ", matchId)

	return &Match{-1}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		presences: make(map[string]runtime.Presence),
	}

	tickRate := 5
	label := ""

	if _, err := CallRPCs["games/create"](ctx, logger, db, nk, "{}"); err != nil {
		logger.Error(fmt.Errorf("Nakama: error creating game", err).Error())
	}

	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {

	_, contains := state.(*MatchState).presences[presence.GetUserId()]// whether user should be accepted

	return state, !contains, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		mState.presences[p.GetUserId()] = p
		result, err := CallRPCs["games/push"](ctx, logger, db, nk, "{\"Name\":\"" + p.GetUserId() + "\"}")

		if err != nil {
			return err
		}

		fmt.Println("player joined: ", p.GetUserId(), "; result: ", result)
	}

	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		result, err := CallRPCs["games/pop"](ctx, logger, db, nk, "{\"Name\":\"" + p.GetUserId() + "\"}")

		if err != nil {
			return err
		}
		
		delete(mState.presences, p.GetUserId())

		fmt.Println("player left: ", p.GetUserId(), "; result: ", result)
	}

	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	// process last player input for each type of input
	messageMap := make(map[string] map[int64] []byte)

	for _, match := range messages {
		if _, contains := mState.presences[match.GetUserId()]; !contains {
			return fmt.Errorf("Nakama: unregistered player is moving")
		}

		messageMap[match.GetUserId()][match.GetOpCode()] = match.GetData()
	}

	for _, matchMap := range messageMap {
		for opCode, matchData := range matchMap {
			var err error

			switch opCode {
				case MOVE: _, err = CallRPCs["games/move"](ctx, logger, db, nk, string(matchData))// the move should contain the player name, so it shouldn't be necessary to also include the presence name in here
			}

			if err != nil {
				return err
			}
		}

	}

	// broadcast
	_, err := CallRPCs["games/tick"](ctx, logger, db, nk, "{}")

	if err != nil {
		return err
	}

	for _, p := range mState.presences {
		playerState, err := CallRPCs["games/status"](ctx, logger, db, nk, "{\"Name\":\"" + p.GetUserId() + "\"}")
		
		if err != nil {
			return err
		}

		err = dispatcher.BroadcastMessage(OK, []byte(playerState), []runtime.Presence{p}, nil, true)// idk what the boolean is for the last argument of BroadcastMessage, but it isn't listed in the docs

		if err != nil {
			return err
		}
	}

	CallRPCs["games/spawncoins"](ctx, logger, db, nk, "{}")// triggers coin spawn

	m.tick++
	
	return mState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state// there is nothing to do on the Cardinal side to shut the game/world down
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}

func makeEndpoint(currEndpoint string, makeURL func(string) string) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
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
	}
	// the currEndpoint in callEndpoint is declared by the parent environment, which is this function
}

func InitializeCardinalProxy(logger runtime.Logger, initializer runtime.Initializer) error {// initializes cardinal endpoints
	gameServerAddr := os.Getenv(EnvGameServer)
	if gameServerAddr == "" {
		msg := fmt.Sprintf("Must specify a game server via %s", EnvGameServer)
		logger.Error(msg)
		return errors.New(msg)
	}

	makeURL := func(resource string) string {
		return fmt.Sprintf("%s/%s", gameServerAddr, resource)
	}

	// get the list of available endpoints from the backend server
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

		// function creates functions to use for calling endpoints within the code
		CallRPCs[currEndpoint] = makeEndpoint(currEndpoint, makeURL)
		err := initializer.RegisterRpc(e, CallRPCs[currEndpoint])
		if err != nil {
			logger.Error("failed to register endpoint %q: %v", e, err)
		}
	}
	return nil
}
