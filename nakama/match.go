package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/heroiclabs/nakama-common/runtime"
	"strconv"
	"time"
)

// MatchState contains match data as the match progresses
type MatchState struct{}

// Match should contain data on the match, but because this is being handled
// through Cardinal, it's an empty struct that does nothing ---
// all match functions call Cardinal endpoints rather than actually
// updating some Nakama match state
type Match struct {
	tick int
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

		var coins int

		logger.Debug(fmt.Sprintf("Nakama: player push JSON:", "{\"Name\":\""+p.GetUserId()+"\",\"Coins\":"+strconv.Itoa(coins)+"}"))
		result, err := CallRPCs["tx-add-player"](ctx, logger, db, nk, "{\"Name\":\""+p.GetUserId()+"\",\"Coins\":0}")

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
		result, err := CallRPCs["tx-remove-player"](ctx, logger, db, nk, "{\"Name\":\""+presences[i].GetUserId()+"\"}")

		if err != nil {
			logger.Debug(fmt.Errorf("Nakama: error popping player:", err).Error())
		}

		err = dispatcher.BroadcastMessage(REMOVE, []byte(presences[i].GetUserId()), nil, nil, true) // broadcast player removal to all players

		if _, contains := Presences[presences[i].GetUserId()]; contains {
			delete(Presences, presences[i].GetUserId())
		}

		fmt.Println("player left: ", presences[i].GetUserId(), "; result: ", result)
	}

	return state
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	// process player input for each type of input
	messageMap := make(map[string]map[int64][][]byte)

	for _, match := range messages {
		if messageMap[match.GetUserId()] == nil {
			messageMap[match.GetUserId()] = make(map[int64][][]byte)
		}

		messageMap[match.GetUserId()][match.GetOpCode()] = append(messageMap[match.GetUserId()][match.GetOpCode()], match.GetData())

		if _, contains := Presences[match.GetUserId()]; !contains {
			return fmt.Errorf("Nakama: unregistered player is moving")
		}
	}

	for _, matchMap := range messageMap {
		for opCode, matchDataArray := range matchMap {
			var err error

			switch opCode {
			case MOVE:
				for _, matchData := range matchDataArray {
					if _, err = CallRPCs["tx-move-player"](ctx, logger, db, nk, string(matchData)); err != nil { // the move should contain the player name, so it shouldn't be necessary to also include the presence name in here
						logger.Error(fmt.Errorf("Nakama: error registering input:", err).Error())
					}

				}
			}

			if err != nil {
				return err
			}
		}
	}

	// get player statuses; if this does not throw an error, broadcast to everyone & offload coins, otherwise add to removal list
	kickList := make([]string, 0)
	for _, pp := range Presences {
		// Check that it's been 500ms since the player joined, before querying for their state
		if joinTimeMap[pp.GetUserId()].Add(time.Millisecond * 500).After(time.Now()) {
			continue
		}
		// get player state
		playerState, err := CallRPCs["read-player-state"](ctx, logger, db, nk, "{\"player_name\":\""+pp.GetUserId()+"\"}")

		if err != nil { // assume that an error here means the player is dead
			kickList = append(kickList, pp.GetUserId())
		} else { // send everyone player state & send player its nearby coins
			err = dispatcher.BroadcastMessage(LOCATION, []byte(playerState), nil, nil, true) // idk what the boolean is for the last argument of BroadcastMessage, but it isn't listed in the docs

			if err != nil {
				return err
			}

			nearbyCoins, err := CallRPCs["read-player-coins"](ctx, logger, db, nk, "{\"player_name\":\""+pp.GetUserId()+"\"}")

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
	attacks, err := CallRPCs["read-attacks"](ctx, logger, db, nk, "{}")

	if err != nil {
		logger.Error(fmt.Errorf("Nakama: error fetching attack information: ", err).Error())
	}

	if attacks != "[]\n" {
		//logger.Debug(fmt.Sprintf("Nakama: attacks: ", attacks))
		if err = dispatcher.BroadcastMessage(ATTACKS, []byte(attacks), nil, nil, true); err != nil {
			return err
		}
	}

	return state
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state // there is nothing to do on the Cardinal side to shut the game/world down
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
