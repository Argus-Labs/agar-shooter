package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

var counter = time.Now().UnixMilli()
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

// Initializes the Nakama match
func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	tickRate := 10
	label := ""

	return MatchState{}, tickRate, label
}

// Checks whether a client that wants to join the game is allowed to join --- called when the client attempts to join the match; if the result is true, MatchJoin is called on the presence
func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	_, contains := Presences[presence.GetUserId()] // whether user should be accepted

	if contains {
		dispatcher.BroadcastMessage(REJECT, []byte(""), []runtime.Presence{presence}, nil, true)
	}

	return MatchState{}, !contains, ""
}

// Called after MatchJoinAttempt to add the player to the Nakama match
func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if presences == nil {
		return fmt.Errorf("Nakama: no presence exists in MatchJoin")
	}

	for _, p := range presences {
		Presences[p.GetUserId()] = p

		// Call tx-create-pesrona to get a persona tag for the player
		if _, err := cardinalCreatePersona(ctx, nk, p.GetUserId()); err != nil {
			return err
		}

		// Call tx-add-player with newly created persona
		_, err := rpcEndpoints["tx-add-player"](ctx, logger, db, nk, "{\"Name\":\""+p.GetUserId()+"\",\"Coins\":0}")

		if err != nil {
			return err
		}

		joinTimeMap[p.GetUserId()] = time.Now()

		// assign name deterministically
		name := ""
		if len(Presences) > len(IDNameArr) {
			logger.Error("Nakama: too many players in the game")
			break
		}
		for i := rollHash(p.GetUserId()) % len(IDNameArr); ; i = (i + 1) % len(IDNameArr) {
			if !NameTakenMap[IDNameArr[i]] {
				name = IDNameArr[i]
				break
			}
		}

		NameTakenMap[name] = true
		NameToNickname[p.GetUserId()] = name
	}

	return MatchState{}
}

// Called when a player is kicked out of the game for dying
func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if Presences == nil {
		return fmt.Errorf("Nakama: no presence exists")
	}

	for i := 0; i < len(presences); i++ {
		_, err := rpcEndpoints["tx-remove-player"](ctx, logger, db, nk, "{\"Name\":\""+presences[i].GetUserId()+"\"}")

		if err != nil {
			logger.Info(fmt.Errorf("Nakama: error popping player:", err).Error())
		}

		// broadcast player removal to all players
		err = dispatcher.BroadcastMessage(REMOVE, []byte(presences[i].GetUserId()), nil, nil, true)

		if _, contains := Presences[presences[i].GetUserId()]; contains {
			delete(Presences, presences[i].GetUserId())
		}

		// nickname stuff
		NameTakenMap[NameToNickname[presences[i].GetUserId()]] = false
		delete(NameToNickname, presences[i].GetUserId())
	}

	return MatchState{}
}

// Processes messages (the list of messages sent from each player to Nakama) and broadcasts necessary information to each player
func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	CardinalOpCounter = 0
	playerInputNum := make(map[string] int)
	playerInputSeqNum := make(map[string] string)

	for _, m := range messages {
		switch m.GetOpCode() {
		case MOVE:
			if key, contains := playerInputNum[m.GetUserId()]; contains && key >= 20 {
				playerInputNum[m.GetUserId()]++
				continue
			}
			
			data := m.GetData()
			if _, err := rpcEndpoints["tx-move-player"](ctx, logger, db, nk, string(data)); err != nil {
				logger.Error(fmt.Errorf("Nakama: error registering input:", err).Error())
			}

			playerInputNum[m.GetUserId()]++
			playerInputSeqNum[m.GetUserId()] = strings.Split(string(data),"Input_sequence_number")[1][2:9]
		}
	}

	// get player statuses; if this does not throw an error, broadcast to everyone & offload coins, otherwise add to removal list
	kickList := make([]string, 0)
	for _, pp := range Presences {
		// Check that it's been a second since the player joined before querying for their state to give Cardinal time to add them and give them a buffer
		if joinTimeMap[pp.GetUserId()].Add(time.Second).After(time.Now()) {
			if gameParameters, err := rpcEndpoints["read-game-parameters"](ctx, logger, db, nk, "{}"); err != nil { // assume that an error here means the player is dead
				return err
			} else { // send everyone player state & send player its nearby coins
				if err = dispatcher.BroadcastMessage(PARAMS, []byte(gameParameters), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
			}
			continue
		}

		// Create request body
		reqBody := PlayerPersonaRequest{PlayerPersona: pp.GetUserId()}
		reqJSON, err := json.Marshal(reqBody)
		req := string(reqJSON)
		if err != nil {
			return err
		}
		// Get player state
		

		if playerState, err := rpcEndpoints["read-player-state"](ctx, logger, db, nk, req); err != nil { // assume that an error here means the player is dead
			kickList = append(kickList, pp.GetUserId())
		} else { // send everyone player state & send player its nearby coins
			if err = dispatcher.BroadcastMessage(LOCATION, []byte(playerState), nil, nil, true); err != nil { // idk what the boolean is for the last argument of BroadcastMessage, but it isn't listed in the docs
				return err
			}

			if nearbyCoins, err := rpcEndpoints["read-player-coins"](ctx, logger, db, nk, req); err != nil {
				return err
			} else {
				if err := dispatcher.BroadcastMessage(COINS, []byte(nearbyCoins), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
			}
			
			
			if nearbyHealth, err := rpcEndpoints["read-player-health"](ctx, logger, db, nk, req); err != nil {
				return err
			} else {
				if err = dispatcher.BroadcastMessage(HEALTH, []byte(nearbyHealth), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
			}
		}
	}

	// kick all dead players
	for _, pid := range kickList {
		if err := dispatcher.BroadcastMessage(DED, []byte(""), []runtime.Presence{Presences[pid]}, nil, true); err != nil {
			return err
		}

		if err := dispatcher.BroadcastMessage(REMOVE, []byte(pid), nil, nil, true); err != nil { // broadcast player removal to all players
			return err
		}

		dispatcher.MatchKick([]runtime.Presence{Presences[pid]})
		delete(Presences, pid)
	}

	// send attack information to all players
	if attacks, err := rpcEndpoints["read-attacks"](ctx, logger, db, nk, "{}"); err != nil {
		logger.Error(fmt.Errorf("Nakama: error fetching attack information: ", err).Error())
	} else {
		if attacks != "[]\n" {
			if err = dispatcher.BroadcastMessage(ATTACKS, []byte(attacks), nil, nil, true); err != nil {
				return err
			}
		}
	}

	if _, err := rpcEndpoints["read-tick"](ctx, logger, db, nk, "{}"); err != nil {
		return fmt.Errorf("Nakama: tick error: %w", err)
	}

	m.tick++
	//broadcast player nicknames
	if len(NameToNickname) > 0 {
		stringmap := "["
		for key, val := range NameToNickname {
			stringmap += "{\"UserId\":\"" + key + "\",\"Name\":\"" + val + "\"},"
		}
		stringmap = stringmap[:len(stringmap)-1] + "]"

		if err := dispatcher.BroadcastMessage(NICKNAME, []byte(stringmap), nil, nil, true); err != nil {
			return err
		}
	}

	return MatchState{}
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state // there is nothing to do on the Cardinal side to shut the game/world down
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
