package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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
		logger.Debug(fmt.Sprint("Nakama: Add Player, JSON:", "{\"PersonaTag\":\""+p.GetUserId()+"\",\"Coins\":0}"))
		result, err := rpcEndpoints["tx-add-player"](ctx, logger, db, nk, "{\"Name\":\""+p.GetUserId()+"\",\"Coins\":0}")

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

		fmt.Println("player joined: ", p.GetUserId(), "; name: ", name, "; result: ", result)
	}

	return MatchState{}
}

// Called when a player is kicked out of the game for dying
func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if Presences == nil {
		return fmt.Errorf("Nakama: no presence exists")
	}

	for i := 0; i < len(presences); i++ {
		result, err := rpcEndpoints["tx-remove-player"](ctx, logger, db, nk, "{\"Name\":\""+presences[i].GetUserId()+"\"}")

		if err != nil {
			logger.Debug(fmt.Errorf("Nakama: error popping player:", err).Error())
		}

		// broadcast player removal to all players
		err = dispatcher.BroadcastMessage(REMOVE, []byte(presences[i].GetUserId()), nil, nil, true)

		if _, contains := Presences[presences[i].GetUserId()]; contains {
			delete(Presences, presences[i].GetUserId())
		}

		fmt.Println("player left: ", presences[i].GetUserId(), "; result: ", result)

		// nickname stuff
		NameTakenMap[NameToNickname[presences[i].GetUserId()]] = false
		delete(NameToNickname, presences[i].GetUserId())
	}

	return MatchState{}
}

// Processes messages (the list of messages sent from each player to Nakama) and broadcasts necessary information to each player
func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	fmt.Printf("MATCH LOOP START\n")
	playerInputNum := make(map[string]int)
	playerInputSeqNum := make(map[string]string)

	diff := time.Now().UnixMilli() - counter
	counter = time.Now().UnixMilli()
	for key, val := range playerInputNum {
		if val >= 20 {
			fmt.Println("Bad player: ", key, NameToNickname[key], strconv.Itoa(val), playerInputSeqNum[key])
		}
	}

	fmt.Println("Nakama MatchLoop (ms):", strconv.Itoa(int(diff)), "\nCardinal Endpoint Calls (ms):", strconv.Itoa(int(CardinalOpCounter/1000)))
	CardinalOpCounter = 0

	t1 := time.Now()
	t2 := time.Now()

	for _, m := range messages {
		switch m.GetOpCode() {
		case MOVE:
			if key, contains := playerInputNum[m.GetUserId()]; contains && key >= 20 {
				playerInputNum[m.GetUserId()]++
				continue
			}

			data := m.GetData()
			t1 = time.Now()
			if _, err := rpcEndpoints["tx-move-player"](ctx, logger, db, nk, string(data)); err != nil {
				logger.Error(fmt.Errorf("Nakama: error registering input:", err).Error())
			}
			t2 = time.Now()
			fmt.Printf("MovePlayerTx: %d\n", t2.Sub(t1).Milliseconds())

			playerInputNum[m.GetUserId()]++
			playerInputSeqNum[m.GetUserId()] = strings.Split(string(data), "Input_sequence_number")[1][2:9]
		}
	}

	// get player statuses; if this does not throw an error, broadcast to everyone & offload coins, otherwise add to removal list
	kickList := make([]string, 0)
	logger.Debug("List of presences in MatchLoop: %v", Presences)
	for _, pp := range Presences {
		// Check that it's been a second since the player joined before querying for their state to give Cardinal time to add them and give them a buffer
		if joinTimeMap[pp.GetUserId()].Add(time.Second).After(time.Now()) {
			t1 = time.Now()
			if gameParameters, err := rpcEndpoints["read-game-parameters"](ctx, logger, db, nk, "{}"); err != nil { // assume that an error here means the player is dead
				return err
			} else { // send everyone player state & send player its nearby coins
				t2 = time.Now()
				fmt.Printf("read-game-parameters: %d\n", t2.Sub(t1).Milliseconds())
				t1 = time.Now()
				if err = dispatcher.BroadcastMessage(PARAMS, []byte(gameParameters), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
				t2 = time.Now()
				fmt.Printf("Broadcast PARAMS: %d\n", t2.Sub(t1).Milliseconds())
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

		t1 = time.Now()
		if playerState, err := rpcEndpoints["read-player-state"](ctx, logger, db, nk, req); err != nil {
			t2 = time.Now()
			fmt.Printf("read-player-state: %d\n", t2.Sub(t1).Milliseconds())
			kickList = append(kickList, pp.GetUserId())
		} else {
			t2 = time.Now()
			fmt.Printf("read-player-state: %d\n", t2.Sub(t1).Milliseconds())
			// send everyone player state & send player its nearby coins
			t1 = time.Now()
			if err = dispatcher.BroadcastMessage(LOCATION, []byte(playerState), nil, nil, true); err != nil { // idk what the boolean is for the last argument of BroadcastMessage, but it isn't listed in the docs
				return err
			}
			t2 = time.Now()
			fmt.Printf("Broadcast LOCATION: %d\n", t2.Sub(t1).Milliseconds())

			t1 = time.Now()
			if nearbyCoins, err := rpcEndpoints["read-player-coins"](ctx, logger, db, nk, req); err != nil {
				return err
			} else {
				t2 = time.Now()
				fmt.Printf("read-player-coins: %d\n", t2.Sub(t1).Milliseconds())
				t1 = time.Now()
				if err := dispatcher.BroadcastMessage(COINS, []byte(nearbyCoins), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
				t2 = time.Now()
				fmt.Printf("Broadcast COINS: %d\n", t2.Sub(t1).Milliseconds())
			}

			t1 = time.Now()
			if nearbyHealth, err := rpcEndpoints["read-player-health"](ctx, logger, db, nk, req); err != nil {
				return err
			} else {
				t2 = time.Now()
				fmt.Printf("read-player-health: %d\n", t2.Sub(t1).Milliseconds())
				t1 = time.Now()
				if err = dispatcher.BroadcastMessage(HEALTH, []byte(nearbyHealth), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
				t2 = time.Now()
				fmt.Printf("Broadcast HEALTH: %d\n", t2.Sub(t1).Milliseconds())
			}

			/*
				if intCoins, err := callRPCs["read-player-totalcoins"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}"); err != nil {
					return err
				} else {
					if err = dispatcher.BroadcastMessage(TOTAL_COINS, []byte(intCoins), nil, nil, true); err != nil {// send coins to all players
						return err
					}
				}
			*/
		}
	}

	// kick all dead players
	for _, pid := range kickList {
		t1 = time.Now()
		if err := dispatcher.BroadcastMessage(DED, []byte(""), []runtime.Presence{Presences[pid]}, nil, true); err != nil {
			return err
		}
		t2 = time.Now()
		fmt.Printf("Broadcast DED: %d\n", t2.Sub(t1).Milliseconds())

		t1 = time.Now()
		if err := dispatcher.BroadcastMessage(REMOVE, []byte(pid), nil, nil, true); err != nil { // broadcast player removal to all players
			return err
		}
		t2 = time.Now()
		fmt.Printf("Broadcast REMOVE: %d\n", t2.Sub(t1).Milliseconds())

		dispatcher.MatchKick([]runtime.Presence{Presences[pid]})
		delete(Presences, pid)
	}

	// send attack information to all players
	t1 = time.Now()
	if attacks, err := rpcEndpoints["read-attacks"](ctx, logger, db, nk, "{}"); err != nil {
		logger.Error(fmt.Errorf("Nakama: error fetching attack information: ", err).Error())
	} else {
		t2 = time.Now()
		fmt.Printf("read-attacks: %d\n", t2.Sub(t1).Milliseconds())
		if attacks != "[]\n" {
			t1 = time.Now()
			if err = dispatcher.BroadcastMessage(ATTACKS, []byte(attacks), nil, nil, true); err != nil {
				return err
			}
			t2 = time.Now()
			fmt.Printf("Broadcast ATTACKS: %d\n", t2.Sub(t1).Milliseconds())
		}
	}

	t1 = time.Now()
	if _, err := rpcEndpoints["read-tick"](ctx, logger, db, nk, "{}"); err != nil {
		return fmt.Errorf("Nakama: tick error: %w", err)
	}
	t2 = time.Now()
	fmt.Printf("read-ticks: %d\n", t2.Sub(t1).Milliseconds())

	m.tick++
	//broadcast player nicknames
	if len(NameToNickname) > 0 {
		stringmap := "["
		for key, val := range NameToNickname {
			stringmap += "{\"UserId\":\"" + key + "\",\"Name\":\"" + val + "\"},"
		}
		stringmap = stringmap[:len(stringmap)-1] + "]"
		t1 = time.Now()
		if err := dispatcher.BroadcastMessage(NICKNAME, []byte(stringmap), nil, nil, true); err != nil {
			return err
		}
		t2 = time.Now()
		fmt.Printf("Broadcast NICKNAME: %d\n", t2.Sub(t1).Milliseconds())
	}
	fmt.Printf("MATCH LOOP END\n")
	return MatchState{}
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state // there is nothing to do on the Cardinal side to shut the game/world down
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
