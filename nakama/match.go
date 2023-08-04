package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

// MatchState contains match data as the match progresses
type MatchState struct{}

// Match should contain data on the match, but because this is being handled
// through Cardinal, it's an empty struct that does nothing --- all match functions
// call Cardinal endpoints rather than actually updating some Nakama match state
type Match struct {
	tick int
}

type PlayerPersonaRequest struct {
	PlayerPersona string `json:"player_persona"`
}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	return &Match{-1}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	tickRate := 10
	label := ""

	time.Sleep(5 * time.Second)

	return MatchState{}, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	_, contains := Presences[presence.GetUserId()] // whether user should be accepted

	return state, !contains, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if presences == nil {
		return fmt.Errorf("Nakama: no presence exists in MatchJoin")
	}

	for _, p := range presences {
		Presences[p.GetUserId()] = p

		// Call tx-create-persona to get a persona tag for the player
		_, err := cardinalCreatePersona(ctx, nk, p.GetUserId())
		if err != nil {
			return err
		}
		// Wait for the persona to be created in cardinal, 200ms = 2 ticks
		time.Sleep(time.Millisecond * 200)

		// Call tx-add-player with newly created persona
		logger.Debug(fmt.Sprint("Nakama: Add Player, JSON:", "{\"PersonaTag\":\""+p.GetUserId()+"\",\"Coins\":0}"))
		result, err := rpcEndpoints["tx-add-player"](ctx, logger, db, nk, "{\"PersonaTag\":\""+p.GetUserId()+"\",\"Coins\":0}")

		if err != nil {
			return err
		}

		joinTimeMap[p.GetUserId()] = time.Now()
		fmt.Println("player joined: ", p.GetUserId(), "; result: ", result)
	}

	return state
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if Presences == nil {
		return fmt.Errorf("Nakama: no presence exists")
	}

	for i := 0; i < len(presences); i++ {
		result, err := rpcEndpoints["tx-remove-player"](ctx, logger, db, nk, "{\"PlayerPersona\":\""+presences[i].GetUserId()+"\"}")

		if err != nil {
			logger.Debug(fmt.Errorf("Nakama: error popping player:", err).Error())
		}

		// broadcast player removal to all players
		err = dispatcher.BroadcastMessage(REMOVE, []byte(presences[i].GetUserId()), nil, nil, true)

		if _, contains := Presences[presences[i].GetUserId()]; contains {
			delete(Presences, presences[i].GetUserId())
		}

		fmt.Println("player left: ", presences[i].GetUserId(), "; result: ", result)
	}

	return state
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	// Map of Player Move messages sent from the game client

	for _, msg := range messages {
		switch msg.GetOpCode() {
		case MOVE:
			data := msg.GetData()
			if _, err := rpcEndpoints["tx-move-player"](ctx, logger, db, nk, string(data)); err != nil {
				logger.Error(fmt.Errorf("Nakama: error registering input:", err).Error())
			}
		}
	}

	// get player statuses; if this does not throw an error, broadcast to everyone & offload coins, otherwise add to removal list
	kickList := make([]string, 0)
	for _, pp := range Presences {
		userID := pp.GetUserId()

		// Check if the user has been created in cardinal already, before querying for its state
		if !isUserIDSafeToQuery(userID, joinTimeMap, isSafeToQueryMap) {
			continue // Skip further processing for this user ID if it's not safe to query
		}

		// Create request body
		reqBody := PlayerPersonaRequest{PlayerPersona: userID}
		reqJSON, err := json.Marshal(reqBody)
		if err != nil {
			return err // Or appropriate error handling
		}
		// Get player state
		playerState, err := rpcEndpoints["read-player-state"](ctx, logger, db, nk, string(reqJSON))

		if err != nil { // assume that an error here means the player is dead
			kickList = append(kickList, userID)
		} else { // send everyone player state & send player its nearby coins
			err = dispatcher.BroadcastMessage(LOCATION, []byte(playerState), nil, nil, true)
			if err != nil {
				return err
			}

			nearbyCoins, err := rpcEndpoints["read-player-coins"](ctx, logger, db, nk, string(reqJSON))
			if err != nil {
				return err
			}

			err = dispatcher.BroadcastMessage(COINS, []byte(nearbyCoins), []runtime.Presence{pp}, nil, true)
			if err != nil {
				return err
			}
		}
	}

	// kick all dead players
	for _, pid := range kickList {
		if err := dispatcher.BroadcastMessage(DED, []byte(""), []runtime.Presence{Presences[pid]}, nil, true); err != nil {
			return err
		}
		dispatcher.MatchKick([]runtime.Presence{Presences[pid]})
		delete(Presences, pid)
	}

	// TODO: @fareed, gotta fix this read-attack stuff
	// send attack information to all players
	//attacks, err := rpcEndpoints["read-attacks"](ctx, logger, db, nk, "{}")
	//
	//if err != nil {
	//	logger.Error(fmt.Errorf("Nakama: error fetching attack information: ", err).Error())
	//}
	//
	//if attacks != "[]\n" {
	//	//logger.Debug(fmt.Sprintf("Nakama: attacks: ", attacks))
	//	if err = dispatcher.BroadcastMessage(ATTACKS, []byte(attacks), nil, nil, true); err != nil {
	//		return err
	//	}
	//}

	return state
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state // there is nothing to do on the Cardinal side to shut the game/world down
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
