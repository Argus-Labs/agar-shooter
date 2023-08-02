package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argus-labs/world-engine/sign"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"math"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	LOCATION int64				= 0
	COINS int64					= 1
	REMOVE int64				= 2
	ATTACKS int64				= 3
	DEADLINE_EXCEEDED int64		= 4
	DED int64					= 5
	TESTADDHEALTH int64			= 6
	EXTRACTION_POINT int64		= 7
	TOTAL_COINS int64			= 8
	NICKNAME int64				= 9
	HEALTH int64				= 10
	OUT_OF_RANGE int64			= 11
	UNIMPLEMENTED int64			= 12
	INTERNAL int				= 13
	UNAVAILABLE int64			= 14
	DATA_LOSS int64				= 15
	UNAUTHENTICATED int64		= 16
	MOVE int64					= 17
)

const (
	EnvGameServer	= "GAME_SERVER_ADDR"
)

var (
	t					= time.Now()
	CallRPCs			= make(map[string] func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error))// contains all RPC endpoint functions
	Presences			= make(map[string] runtime.Presence)// contains all in-game players; stopped storing in a MatchState struct because Nakama throws stupid errors for things that shouldn't happen when I do and checking all these errors is a waste of time
	joinTimeMap			= make(map[string]time.Time) // mapping player id to time they joined
	IDNameArr			= []string {
		"Alice",
		"Bob",
		"Charlie",
		"David",
		"Emma",
		"Frank",
		"Grace",
		"Henry",
		"Isabella",
		"Jack",
		"Kate",
		"Liam",
		"Mia",
		"Noah",
		"Olivia",
		"Paul",
		"Quinn",
		"Ryan",
		"Sophia",
		"Tom",
		"Uma",
		"Vincent",
		"Willow",
		"Xander",
		"Yara",
		"Zara",
		"Adam",
		"Benjamin",
		"Chloe",
		"Dylan",
		"Eva",
		"Finn",
		"Georgia",
		"Hannah",
		"Isaac",
		"Julia",
		"Kai",
		"Lily",
		"Matthew",
		"Nora",
		"Owen",
		"Penelope",
		"Quincy",
		"Riley",
		"Sofia",
		"Tucker",
		"Ursula",
		"Victor",
		"Willa",
		"Xenia",
		"Yasmine",
		"Zoe",
		"Andrew",
		"Bella",
		"Caleb",
		"Daniel",
		"Emily",
		"Faith",
		"Gabriel",
		"Hazel",
		"Ian",
		"Jacob",
		"Katherine",
		"Leo",
		"Michael",
		"Nathan",
		"Oliver",
		"Patrick",
		"Quentin",
		"Rachel",
		"Samuel",
		"Thomas",
		"Ulysses",
		"Victoria",
		"William",
		"Xavier",
		"Yvette",
		"Zachary",
		"Amelia",
		"Ben",
		"Charlotte",
		"Eleanor",
	}
	NameToNickname		= make(map[string] string)
	rollHash			= func(s string) int {
		hash := 0
		for i, c := range s {
			hash = (hash + int(c)*int(math.Pow(26, float64(i%10))))%(1e9+7)
		}

		return hash
	}
	NameTakenMap		= map[string] bool {
		"Alice": false,
		"Bob": false,
		"Charlie": false,
		"David": false,
		"Emma": false,
		"Frank": false,
		"Grace": false,
		"Henry": false,
		"Isabella": false,
		"Jack": false,
		"Kate": false,
		"Liam": false,
		"Mia": false,
		"Noah": false,
		"Olivia": false,
		"Paul": false,
		"Quinn": false,
		"Ryan": false,
		"Sophia": false,
		"Tom": false,
		"Uma": false,
		"Vincent": false,
		"Willow": false,
		"Xander": false,
		"Yara": false,
		"Zara": false,
		"Adam": false,
		"Benjamin": false,
		"Chloe": false,
		"Dylan": false,
		"Eva": false,
		"Finn": false,
		"Georgia": false,
		"Hannah": false,
		"Isaac": false,
		"Julia": false,
		"Kai": false,
		"Lily": false,
		"Matthew": false,
		"Nora": false,
		"Owen": false,
		"Penelope": false,
		"Quincy": false,
		"Riley": false,
		"Sofia": false,
		"Tucker": false,
		"Ursula": false,
		"Victor": false,
		"Willa": false,
		"Xenia": false,
		"Yasmine": false,
		"Zoe": false,
		"Andrew": false,
		"Bella": false,
		"Caleb": false,
		"Daniel": false,
		"Emily": false,
		"Faith": false,
		"Gabriel": false,
		"Hazel": false,
		"Ian": false,
		"Jacob": false,
		"Katherine": false,
		"Leo": false,
		"Michael": false,
		"Nathan": false,
		"Oliver": false,
		"Patrick": false,
		"Quentin": false,
		"Rachel": false,
		"Samuel": false,
		"Thomas": false,
		"Ulysses": false,
		"Victoria": false,
		"William": false,
		"Xavier": false,
		"Yvette": false,
		"Zachary": false,
		"Amelia": false,
		"Ben": false,
		"Charlotte": false,
		"Eleanor": false,
    };

)

type MatchState struct{} // contains match data as the match progresses

type Match struct {
	tick int
} // should contain data on the match, but because this is being handled through Cardinal, it's an empty struct that does nothing --- all match functions call Cardinal endpoints rather than actually updating some Nakama match state

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error { // called when connection is established
	if err := InitializeCardinalProxy(logger, initializer); err != nil { // Register the RPC function of Cardinal to Nakama to create a proxy
		return err
	}

	// Create the singleton match
	if err := initializer.RegisterMatch("singleton_match", newMatch); err != nil {
		return err
	}

	if _, err := nk.MatchCreate(ctx, "singleton_match", map[string]interface{}{}); err != nil { // calls the newMatch function, then calls MatchInit on the result
		return err
	}

	return nil
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

	return MatchState{}, !contains, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if presences == nil {
		return fmt.Errorf("Nakama: no presence exists in MatchJoin")
	}
	logger.Debug("List of presences in Match Join: %+v", presences)

	for _, p := range presences {
		Presences[p.GetUserId()] = p

		result, err := CallRPCs["tx-add-player"](ctx, logger, db, nk, "{\"Name\":\""+p.GetUserId()+"\",\"Coins\":0}")

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
		for i := rollHash(p.GetUserId())%len(IDNameArr);; i = (i+1)%len(IDNameArr) {
			if !NameTakenMap[IDNameArr[i]] {
				name = IDNameArr[i]
				break
			}
		}

		NameTakenMap[name] = true
		NameToNickname[p.GetUserId()] = name

		fmt.Println("player joined: ", p.GetUserId(), "; name: ", name,  "; result: ", result)
	}

	return MatchState{}
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

		// nickname stuff
		NameTakenMap[NameToNickname[presences[i].GetUserId()]] = false
		delete(NameToNickname, presences[i].GetUserId())
	}

	return MatchState{}
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
	logger.Debug("List of presences in MatchLoop: %v", Presences)
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
			if err = dispatcher.BroadcastMessage(LOCATION, []byte(playerState), nil, nil, true); err != nil {// idk what the boolean is for the last argument of BroadcastMessage, but it isn't listed in the docs
				return err
			}

			if nearbyCoins, err := CallRPCs["read-player-coins"](ctx, logger, db, nk, "{\"player_name\":\""+pp.GetUserId()+"\"}"); err != nil {
				return err
			} else {
				if err := dispatcher.BroadcastMessage(COINS, []byte(nearbyCoins), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
			}

			if nearbyHealth, err := CallRPCs["read-player-health"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}"); err != nil {
				return err
			} else {
				if err = dispatcher.BroadcastMessage(HEALTH, []byte(nearbyHealth), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
			}

			if intCoins, err := CallRPCs["read-player-totalcoins"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}"); err != nil {
				return err
			} else {
				if err = dispatcher.BroadcastMessage(TOTAL_COINS, []byte(intCoins), nil, nil, true); err != nil {// send coins to all players
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

		if err := dispatcher.BroadcastMessage(REMOVE, []byte(pid), nil, nil, true); err != nil {// broadcast player removal to all players
			return err
		}

		dispatcher.MatchKick([]runtime.Presence{Presences[pid]})
		delete(Presences, pid)
	}

	// TODO: @fareed, gotta fix this read-attack stuff
	// send attack information to all players
	if attacks, err := CallRPCs["read-attacks"](ctx, logger, db, nk, "{}"); err != nil {
		logger.Error(fmt.Errorf("Nakama: error fetching attack information: ", err).Error())
	} else {
		if attacks != "[]\n" {
			//logger.Debug(fmt.Sprintf("Nakama: attacks: ", attacks))
			if err = dispatcher.BroadcastMessage(ATTACKS, []byte(attacks), nil, nil, true); err != nil {
				return err
			}
		}
	}

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
	return MatchState{} // there is nothing to do on the Cardinal side to shut the game/world down
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return MatchState{}, ""//, "signal received: " + data
}

func makeTxEndpoint(currEndpoint string, makeURL func(string) string) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

		signedPayload := &sign.SignedPayload{
			PersonaTag: "sender's tag",
			Namespace:  "0",
			Nonce:      1000,
			Signature:  "pk",
			Body:       json.RawMessage(payload),
		}

		logger.Debug("Got request for %q, payload: %s", currEndpoint, payload)
		payloadStr, err := json.Marshal(signedPayload)
		//logger.Debug("PayloadStr: %v ", payloadStr)
		readerStr := strings.NewReader(string(payloadStr))
		logger.Debug("string from Reader: %s", readerStr)

		req, err := http.NewRequestWithContext(ctx, "GET", makeURL(currEndpoint), strings.NewReader(string(payloadStr)))
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

func makeReadEndpoint(currEndpoint string, makeURL func(string) string) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

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

func InitializeCardinalProxy(logger runtime.Logger, initializer runtime.Initializer) error { // initializes cardinal endpoints
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
	resp, err := http.Get(makeURL("list/tx-endpoints"))
	if err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	var endpoints []string
	if err := dec.Decode(&endpoints); err != nil {
		return err
	}

	// get the list of available endpoints from the backend server
	resp, err = http.Get(makeURL("list/read-endpoints"))
	if err != nil {
		return err
	}

	dec = json.NewDecoder(resp.Body)
	var endpoints2 []string
	if err := dec.Decode(&endpoints2); err != nil {
		return err
	}

	endpoints = append(endpoints, endpoints2...)

	for _, e := range endpoints {
		endpoint := e
		endpoint = endpoint[1:]
		logger.Debug("registering: %v", endpoint)

		// function creates functions to use for calling endpoints within the code
		if strings.HasPrefix(endpoint, "read") {
			CallRPCs[endpoint] = makeReadEndpoint(endpoint, makeURL)
		} else if strings.HasPrefix(endpoint, "tx") {
			CallRPCs[endpoint] = makeTxEndpoint(endpoint, makeURL)
		} else {
			logger.Error("The following endpoint does not have the correct format: %s", endpoint)
		}
		err := initializer.RegisterRpc(endpoint, CallRPCs[endpoint])
		if err != nil {
			logger.Error("failed to register endpoint %q: %v", endpoint, err)
		}
	}

	return nil
}
