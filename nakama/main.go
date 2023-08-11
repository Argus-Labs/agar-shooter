package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"math"

	"github.com/argus-labs/world-engine/sign"
	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	LOCATION			int64 = 0
	COINS				int64 = 1
	REMOVE				int64 = 2
	ATTACKS				int64 = 3
	DEADLINE_EXCEEDED	int64 = 4
	DED					int64 = 5
	TESTADDHEALTH		int64 = 6
	EXTRACTION_POINT	int64 = 7
	TOTAL_COINS			int64 = 8
	NICKNAME			int64 = 9
	HEALTH				int64 = 10
	REJECT				int64 = 11
	UNIMPLEMENTED		int64 = 12
	INTERNAL			int   = 13
	PARAMS				int64 = 14
	DATA_LOSS			int64 = 15
	UNAUTHENTICATED		int64 = 16
	MOVE				int64 = 17
)

const (
	EnvCardinalAddr = "GAME_SERVER_ADDR"
)

var (
	rpcEndpoints    = make(map[string]func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error)) // contains all RPC endpoint functions
	Presences   = make(map[string]runtime.Presence)                                                                                                      // contains all in-game players; stopped storing in a MatchState struct because Nakama throws stupid errors for things that shouldn't happen when I do and checking all these errors is a waste of time
	joinTimeMap = make(map[string]time.Time)                                                                                                             // mapping player id to time they joined
	IDNameArr   = []string{
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
	NameToNickname = make(map[string]string)
	rollHash       = func(s string) int {
		hash := 0
		for i, c := range s {
			hash = (hash + int(c)*int(math.Pow(26, float64(i%10)))) % (1e9 + 7)
		}

		return hash
	}
	NameTakenMap = map[string]bool{
		"Alice":     false,
		"Bob":       false,
		"Charlie":   false,
		"David":     false,
		"Emma":      false,
		"Frank":     false,
		"Grace":     false,
		"Henry":     false,
		"Isabella":  false,
		"Jack":      false,
		"Kate":      false,
		"Liam":      false,
		"Mia":       false,
		"Noah":      false,
		"Olivia":    false,
		"Paul":      false,
		"Quinn":     false,
		"Ryan":      false,
		"Sophia":    false,
		"Tom":       false,
		"Uma":       false,
		"Vincent":   false,
		"Willow":    false,
		"Xander":    false,
		"Yara":      false,
		"Zara":      false,
		"Adam":      false,
		"Benjamin":  false,
		"Chloe":     false,
		"Dylan":     false,
		"Eva":       false,
		"Finn":      false,
		"Georgia":   false,
		"Hannah":    false,
		"Isaac":     false,
		"Julia":     false,
		"Kai":       false,
		"Lily":      false,
		"Matthew":   false,
		"Nora":      false,
		"Owen":      false,
		"Penelope":  false,
		"Quincy":    false,
		"Riley":     false,
		"Sofia":     false,
		"Tucker":    false,
		"Ursula":    false,
		"Victor":    false,
		"Willa":     false,
		"Xenia":     false,
		"Yasmine":   false,
		"Zoe":       false,
		"Andrew":    false,
		"Bella":     false,
		"Caleb":     false,
		"Daniel":    false,
		"Emily":     false,
		"Faith":     false,
		"Gabriel":   false,
		"Hazel":     false,
		"Ian":       false,
		"Jacob":     false,
		"Katherine": false,
		"Leo":       false,
		"Michael":   false,
		"Nathan":    false,
		"Oliver":    false,
		"Patrick":   false,
		"Quentin":   false,
		"Rachel":    false,
		"Samuel":    false,
		"Thomas":    false,
		"Ulysses":   false,
		"Victoria":  false,
		"William":   false,
		"Xavier":    false,
		"Yvette":    false,
		"Zachary":   false,
		"Amelia":    false,
		"Ben":       false,
		"Charlotte": false,
		"Eleanor":   false,
	}
	nakamaPersonaTag = "nakama-persona"
	globalNamespace = "agar-shooter"
	CardinalOpCounter int64
)

// Initializes all functions and variables for Nakama: fetches all Cardinal endpoints, creates a Nakama match using the MatchCreate function, which calls newMatch, then passes the result to MatchInit
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error { // called when connection is established
	time.Sleep(1*time.Second) // This is necessary to allow Cardinal to load before Nakama tries to fetch the list of endpoints
	if err := initCardinalAddress(); err != nil {
		return err
	}	

	if err := initCardinalEndpoints(logger, initializer); err != nil { // Register the RPC function of Cardinal to Nakama to create a proxy
		return err
	}

	if err := initializer.RegisterMatch("singleton_match", newMatch); err != nil {
		return err
	}

	if _, err := nk.MatchCreate(ctx, "singleton_match", map[string]interface{}{}); err != nil { // calls the newMatch function, then calls MatchInit on the result
		return err
	}

	return nil
}

// Returns a function that takes a request and sends it to the given endpoint
func makeEndpoint(currEndpoint string, makeURL func(string) string) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		diffTime := time.Now().UnixMicro()
		logger.Debug("Got request for %q, with payload: %q", currEndpoint, payload)

		signedPayload, err := makeSignedPayload(ctx, nk, payload)
		if err != nil {
			return logError(logger, "unable to make signed payload: %v", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", makeURL(currEndpoint), signedPayload)
		if err != nil {
			return logError(logger, "request setup failed for endpoint %q: %v", currEndpoint, err)
		}
		
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return logError(logger, "request failed for endpoint %q: %v", currEndpoint, err)
		}
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			return logError(logger, "bad status code: %v: %s", resp.Status, body)
		}
		
		str, err := io.ReadAll(resp.Body)
		if err != nil {
			return logError(logger, "can't read body: %v", err)
		}
		CardinalOpCounter += time.Now().UnixMicro() - diffTime

		return string(str), nil
	}
}

// initCardinalEndpoints queries the cardinal server to find the list of existing endpoint, and attempts to set up RPC wrappers around each one.
func initCardinalEndpoints(logger runtime.Logger, initializer runtime.Initializer) error {
	endpoints, err := cardinalListAllEndpoints()

	if err != nil {
		return fmt.Errorf("failed to get list of cardinal endpoints: %w", err)
	}

	for _, e := range endpoints {
		logger.Debug("registering: %v", e)
		currEndpoint := e
	
		if currEndpoint[0] == '/' {
			currEndpoint = currEndpoint[1:]
		}
	
		rpcEndpoints[currEndpoint] = makeEndpoint(currEndpoint, makeURL)
		err := initializer.RegisterRpc(currEndpoint, rpcEndpoints[e])
		if err != nil {
			return err
		}
	}

	return nil
}

// Converts a payload/message into a signed payload
func makeSignedPayload(ctx context.Context, nk runtime.NakamaModule, payload string) (io.Reader, error) {
	personaTag := nakamaPersonaTag

	// pk, nonce, err := getPrivateKeyAndANonce(ctx, nk)
	sp := &sign.SignedPayload {
		PersonaTag: personaTag,
		Namespace: globalNamespace,
		Nonce: 1000,
		Signature: "pk",
		Body: json.RawMessage(payload),
	}

	buf, err := json.Marshal(sp)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf), nil
}

// Debug function
func logCode(logger runtime.Logger, code int, format string, v ...interface{}) (string, error) {
	err := fmt.Errorf(format, v...)
	logger.Error(err.Error())

	return "", runtime.NewError(err.Error(), code)
}

// Debug function
func logError(logger runtime.Logger, format string, v ...interface{}) (string, error) {
	err := fmt.Errorf(format, v...)
	logger.Error(err.Error())

	return "", runtime.NewError(err.Error(), INTERNAL)
}
