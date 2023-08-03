package main

import (
	"context"
	"database/sql"
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

type PlayerMessage struct {
	UserID string
	OpCode int64
	Data   [][]byte
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
		// Wait for the persona to be created in cardinal
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
	messageMap := make(map[string]map[int64][][]byte)

	for _, match := range messages {
		userID := match.GetUserId() // this is the player's persona tag
		opCode := match.GetOpCode() // this is the operation code for the message

		// Check for unregistered player
		if _, contains := Presences[userID]; !contains {
			return fmt.Errorf("Nakama: unregistered player is moving")
		}

		if _, exists := messageMap[userID]; !exists {
			messageMap[userID] = make(map[int64][][]byte)
		}

		// The data represents the actual message sent from the client
		messageMap[userID][opCode] = append(messageMap[userID][opCode], match.GetData())
	}

	for _, matchMap := range messageMap {
		for opCode, matchDataArray := range matchMap {
			switch opCode {
			case MOVE:
				for _, matchData := range matchDataArray {
					if _, err := rpcEndpoints["tx-move-player"](ctx, logger, db, nk, string(matchData)); err != nil {
						logger.Error("Nakama: error registering input:", err)
						return err
					}
				}
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
		playerState, err := rpcEndpoints["read-player-state"](ctx, logger, db, nk, "{\"player_persona\":\""+pp.GetUserId()+"\"}")

		if err != nil { // assume that an error here means the player is dead
			kickList = append(kickList, pp.GetUserId())
		} else { // send everyone player state & send player its nearby coins
			err = dispatcher.BroadcastMessage(LOCATION, []byte(playerState), nil, nil, true)
			if err != nil {
				return err
			}

			nearbyCoins, err := rpcEndpoints["read-player-coins"](ctx, logger, db, nk, "{\"player_persona\":\""+pp.GetUserId()+"\"}")
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
