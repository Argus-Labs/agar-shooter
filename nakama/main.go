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
	"time"
	"strconv"
	"regexp"
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
	MAX_COINS int64				= 8
	NICKNAME int64				= 9
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
	EnvGameServer	= "GAME_SERVER_ADDR"
)

var (
	CallRPCs		= make(map[string] func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error))// contains all RPC endpoint functions
	Presences		= make(map[string] runtime.Presence)// contains all in-game players; stopped storing in a MatchState struct because Nakama throws stupid errors for things that shouldn't happen when I do and checking all these errors is a waste of time
	nonnum			= regexp.MustCompile(`[^0-9]`)
	IDNameArr		= []string{
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
	rollHash		= func(s string) int {
		hash := 0
		for i, c := range s {
			hash = (hash + int(c)*int(math.Pow(26, float64(i%10))))%(1e9+7)
		}

		return hash
	}
	NameTakenMap	= map[string] bool {
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

type DBPlayer struct {
	ID				string
	StoredCoins, CurrCoins 	int
	// add weapon & other base information
}

type MatchState struct {}// contains match data as the match progresses

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

	if _, err := nk.MatchCreate(ctx, "singleton_match", map[string]interface{}{}); err != nil {// calls the newMatch function, then calls MatchInit on the result
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

	if _, err := CallRPCs["games/create"](ctx, logger, db, nk, "{}"); err != nil {
		logger.Error(fmt.Errorf("Nakama: error creating game", err).Error())
	}

	time.Sleep(5*time.Second)

	if _, err := db.Query("DROP TABLE IF EXISTS dbplayer"); err != nil {
		logger.Error(fmt.Errorf("Nakama: error removing all tables", err).Error())// drop table and replace with new table
	}

	if _, err := db.Query("CREATE TABLE dbplayer (id text, storedcoins int, currcoins int)"); err != nil {
		logger.Error(fmt.Errorf("Nakama: error creating table: ", err).Error())
	} else {
		logger.Debug("Nakama: initialized Postgres table")
	}

	return MatchState{}, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	_, contains := Presences[presence.GetUserId()]// whether user should be accepted

	return MatchState{}, !contains, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if presences == nil {
		return fmt.Errorf("Nakama: no presence exists in MatchJoin")
	}

	for _, p := range presences {
		Presences[p.GetUserId()] = p

		// fetch any player information from db
		pingErr := db.Ping()
		if pingErr != nil {
			logger.Error(fmt.Errorf("Nakama: error accessing database: ", pingErr).Error())
		}

		fmt.Println("Connected to database")

		var (
			coins int
			discard1, discard2 string
		)

		row := db.QueryRow("SELECT * FROM dbplayer WHERE id = $1", p.GetUserId())

		if err := row.Scan(&discard1, &coins, &discard2); err != nil {
			if err == sql.ErrNoRows {
				logger.Debug("Nakama: player does not already exist in database; adding player to database")
				
				_, err := db.Exec("INSERT INTO dbplayer (id, storedcoins, currcoins) VALUES ($1, $2, $3)", p.GetUserId(), 0, 0)
				if err != nil {
					logger.Error(fmt.Errorf("Nakama: error inserting player into database: ", err).Error())
				}

				coins = 0
			} else {
				logger.Error(fmt.Errorf("Nakama: error querying prior player information: ", err).Error())
				return state
			}
		} else {
			logger.Debug(fmt.Sprintf("Nakama: player exists in database with coins", coins))
		}
		
		// send player database information to Cardinal if it exists when initializing player
		logger.Debug(fmt.Sprintf("Nakama: player push JSON:", "{\"Name\":\"" + p.GetUserId() +  "\",\"Coins\":" + strconv.Itoa(coins) + "}"))
		_, err := CallRPCs["games/push"](ctx, logger, db, nk, "{\"Name\":\"" + p.GetUserId() +  "\",\"Coins\":0}")

		if err != nil {
			return err
		}

		// assign name deterministically
		name := ""
		if len(Presences) > len(IDNameArr) {
			logger.Error("Nakama: too many players in the game")
			break
		}
		for i := rollHash(p.GetUserId())%len(IDNameArr); ; i = (i+1)%len(IDNameArr) {
			if !NameTakenMap[IDNameArr[i]] {
				name = IDNameArr[i]
				break
			}
		}

		NameTakenMap[name] = true
		NameToNickname[p.GetUserId()] = name

		fmt.Println("player joined: ", p.GetUserId(), "; name: ", name)
	}

	return MatchState{}
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	if Presences == nil {
		return fmt.Errorf("Nakama: no presence exists")
	}

	for i := 0; i < len(presences); i++ {
		result, err := CallRPCs["games/pop"](ctx, logger, db, nk, "{\"Name\":\"" + presences[i].GetUserId() + "\"}")

		if err != nil {
			logger.Debug(fmt.Errorf("Nakama: error popping player:", err).Error())
		}
		
		err = dispatcher.BroadcastMessage(REMOVE, []byte(presences[i].GetUserId()), nil, nil, true)// broadcast player removal to all players

		if _, contains := Presences[presences[i].GetUserId()]; contains {
			delete(Presences, presences[i].GetUserId())
		}

		fmt.Println("player left: ", presences[i].GetUserId(), "; result: ", result)

		// nickname stuff
		NameTakenMap[NameToNickname[presences[i].GetUserId()]] = false;
		delete(NameToNickname, presences[i].GetUserId())
	}

	return MatchState{}
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	// process player input for each type of input
	messageMap := make(map[string] map[int64] [][]byte)

	for _, match := range messages {
		if messageMap[match.GetUserId()] == nil {
			messageMap[match.GetUserId()] = make(map[int64] [][]byte)
		}

		messageMap[match.GetUserId()][match.GetOpCode()] = append(messageMap[match.GetUserId()][match.GetOpCode()], match.GetData())

		if _, contains := Presences[match.GetUserId()]; !contains {
			return fmt.Errorf("Nakama: unregistered player is moving")
		}

		if match.GetOpCode() == TESTADDHEALTH {
			if _, err := CallRPCs["games/testaddhealth"](ctx, logger, db, nk, "{\"Name\":\"" + match.GetUserId() + "\"}"); err != nil {
				logger.Error(fmt.Errorf("Nakama: error adding health:", err).Error())
			}
		}
	}

	for _, matchMap := range messageMap {
		for opCode, matchDataArray := range matchMap {
			var err error

			switch opCode {
				case MOVE:
					for _, matchData := range matchDataArray {// because this just adds the move to the transaction queue, if a handling error occurs here, it's because lots of inputs are processed at once nakama-side. the RPC call itself is very efficient, so nakama likely cannot handle all of the inputs being received from the client at the same time
						if _, err = CallRPCs["games/move"](ctx, logger, db, nk, string(matchData)); err != nil {// the move should contain the player name, so it shouldn't be necessary to also include the presence name in here
							logger.Error(fmt.Errorf("Nakama: error registering input:", err).Error())
						}
					}
			}

			if err != nil {
				return err
			}
		}
	}

	// call tick (could cause players to die, but we're fine as long as we check immediately after), then get player statuses and broadcast to everyone & offload coins; after this, send attack information to all existing players
	// tick
	if _, err := CallRPCs["games/tick"](ctx, logger, db, nk, "{}"); err != nil {
		return fmt.Errorf("Nakama: tick error: %w", err)
	}

	m.tick++

	// get player statuses; if this does not throw an error, broadcast to everyone & offload coins, otherwise add to removal list
	kickList := make([]string, 0)
	for _, pp := range Presences {
		// get player state
		playerState, err := CallRPCs["games/state"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}")
		
		if err != nil {// assume that an error here means the player is dead
			kickList = append(kickList, pp.GetUserId())	
		} else {// send everyone player state & send player its nearby coins
			if err = dispatcher.BroadcastMessage(LOCATION, []byte(playerState), nil, nil, true); err != nil {// idk what the boolean is for the last argument of BroadcastMessage, but it isn't listed in the docs
				return err
			}

			if nearbyCoins, err := CallRPCs["games/coins"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}"); err != nil {
				return err
			} else {
				if err = dispatcher.BroadcastMessage(COINS, []byte(nearbyCoins), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
			}

			if extractionPoint, err := CallRPCs["games/extract"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}"); err != nil {
				return err
			} else {
				if err = dispatcher.BroadcastMessage(EXTRACTION_POINT, []byte(extractionPoint), []runtime.Presence{pp}, nil, true); err != nil {
					return err
				}
			}

			if intCoins, err := CallRPCs["games/offload"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}"); err != nil {
				return err
			} else {
				if err = dispatcher.BroadcastMessage(MAX_COINS, []byte(intCoins), nil, nil, true); err != nil {// send max coins to all players
					return err
				}
			}
		}
		
		// offload coins to database
		coins, err := CallRPCs["games/offload"](ctx, logger, db, nk, "{\"Name\":\"" + pp.GetUserId() + "\"}")
		var intCoins int
	
		if err != nil {
			logger.Error(fmt.Errorf("Nakama: error fetching extraction point stuff: ", err).Error())
		}

		if intCoins, err = strconv.Atoi(strings.TrimSpace(coins)); err != nil {
			logger.Error(fmt.Errorf("Nakama:", err).Error())
		}

		if intCoins > 0 {
			var (
				coins int
				discard1, discard2 string
			)

			row := db.QueryRow("SELECT * FROM dbplayer WHERE id = $1", pp.GetUserId())

			if err := row.Scan(&discard1, &coins, &discard2); err != nil {
				if err != sql.ErrNoRows && err != nil {
					logger.Error(fmt.Errorf("Nakama: error getting player information from database: ", err).Error())
					return state
				} else {
					logger.Error(fmt.Errorf("Nakama: error querying prior player information: ", err).Error())
					return state
				}
			}
		
			if _, err = db.Exec("UPDATE dbplayer SET storedcoins = $1 WHERE id = $2", intCoins + coins, pp.GetUserId()); err != nil {
				logger.Error(fmt.Errorf("Nakama: error updating player's coins: ", err).Error())
			} else {
				logger.Debug(fmt.Sprintf("Nakama: player's coins updated: ", intCoins + coins))
			}
		}
	}

	// kick all dead players
	var (
		coins int
		discard1, discard2 string
	)
	for _, pid := range kickList {
		row := db.QueryRow("SELECT * FROM dbplayer WHERE id = $1", pid)

		if err := row.Scan(&discard1, &coins, &discard2); err != nil {
			if err != sql.ErrNoRows && err != nil {
				logger.Error(fmt.Errorf("Nakama: error getting player information from database: ", err).Error())
			} else {
				logger.Error(fmt.Errorf("Nakama: error querying prior player information: ", err).Error())
			}
			coins = -1
		}

		if err := dispatcher.BroadcastMessage(DED, []byte(strconv.Itoa(coins)), []runtime.Presence{Presences[pid]}, nil, true); err != nil {
			return err
		}

		if err := dispatcher.BroadcastMessage(REMOVE, []byte(pid), nil, nil, true); err != nil {// broadcast player removal to all players
			return err
		}

		dispatcher.MatchKick([]runtime.Presence{Presences[pid]})
		delete(Presences, pid)
	}


	// send attack information to all players
	attacks, err := CallRPCs["games/attacks"](ctx, logger, db, nk, "{}")

	if err != nil {
		logger.Error(fmt.Errorf("Nakama: error fetching attack information: ", err).Error())
	}

	if attacks != "[]\n" {
		//logger.Debug(fmt.Sprintf("Nakama: attacks: ", attacks))
		if err = dispatcher.BroadcastMessage(ATTACKS, []byte(attacks), nil, nil, true); err != nil {
			return err
		}
	}

	// broadcast player nicknames
	if len(NameToNickname) > 0 {
		stringmap := "["
		for key, val := range NameToNickname {
			stringmap += "{\"UserId\":\"" + key + "\",\"Name\":\"" + val + "\"},"
		}
		stringmap += "]"
		
		if err := dispatcher.BroadcastMessage(NICKNAME, []byte(stringmap), nil, nil, true); err != nil {
			return err
		}
	}
	
	return MatchState{}
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return MatchState{}// there is nothing to do on the Cardinal side to shut the game/world down
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return MatchState{}, ""//"signal received: " + data
}

func makeEndpoint(currEndpoint string, makeURL func(string) string) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		if currEndpoint == "games/push" || currEndpoint == "games/pop" { logger.Info("Got request for %q", currEndpoint) }

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
