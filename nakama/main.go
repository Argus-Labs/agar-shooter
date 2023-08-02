package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/argus-labs/world-engine/sign"
	"github.com/heroiclabs/nakama-common/runtime"
	"io"
	"net/http"
	"time"
)

const (
	LOCATION            int64 = 0
	COINS               int64 = 1
	REMOVE              int64 = 2
	ATTACKS             int64 = 3
	DEADLINE_EXCEEDED   int64 = 4
	DED                 int64 = 5
	TESTADDHEALTH       int64 = 6
	PERMISSION_DENIED   int64 = 7
	RESOURCE_EXHAUSTED  int64 = 8
	FAILED_PRECONDITION int64 = 9
	ABORTED             int64 = 10
	OUT_OF_RANGE        int64 = 11
	UNIMPLEMENTED       int64 = 12
	INTERNAL            int   = 13
	UNAVAILABLE         int64 = 14
	DATA_LOSS           int64 = 15
	UNAUTHENTICATED     int64 = 16
	MOVE                int64 = 17
)

const (
	EnvCardinalAddr = "GAME_SERVER_ADDR"
)

var (
	CallRPCs         = make(map[string]func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error)) // contains all RPC endpoint functions
	Presences        = make(map[string]runtime.Presence)                                                                                                      // contains all in-game players; stopped storing in a MatchState struct because Nakama throws stupid errors for things that shouldn't happen when I do and checking all these errors is a waste of time
	joinTimeMap      = make(map[string]time.Time)                                                                                                             // mapping player id to time they joined
	nakamaPersonaTag = "nakama-persona"
	globalNamespace  = "agar-shooter"
)

type DBPlayer struct {
	ID                     string
	StoredCoins, CurrCoins int
	// add weapon & other base information
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error { // called when connection is established
	if err := initCardinalAddress(); err != nil {
		return err
	}

	if err := initCardinalEndpoints(logger, initializer); err != nil {
		return fmt.Errorf("failed to init cardinal endpoints: %w", err)
	}

	if err := initializer.RegisterMatch("singleton_match", newMatch); err != nil {
		return err
	}

	if _, err := nk.MatchCreate(ctx, "singleton_match", map[string]interface{}{}); err != nil { // calls the newMatch function, then calls MatchInit on the result
		return err
	}

	return nil
}

// initCardinalEndpoints queries the cardinal server to find the list of existing endpoints, and attempts to
// set up RPC wrappers around each one.
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
		err := initializer.RegisterRpc(currEndpoint, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
			logger.Debug("Got request for %q", currEndpoint)

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
			return string(str), nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func makeSignedPayload(ctx context.Context, nk runtime.NakamaModule, payload string) (io.Reader, error) {
	personaTag := nakamaPersonaTag

	pk, nonce, err := getPrivateKeyAndANonce(ctx, nk)
	sp, err := sign.NewSignedString(pk, personaTag, globalNamespace, nonce, payload)
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(sp)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

func logCode(logger runtime.Logger, code int, format string, v ...interface{}) (string, error) {
	err := fmt.Errorf(format, v...)
	logger.Error(err.Error())
	return "", runtime.NewError(err.Error(), code)
}

func logError(logger runtime.Logger, format string, v ...interface{}) (string, error) {
	err := fmt.Errorf(format, v...)
	logger.Error(err.Error())
	return "", runtime.NewError(err.Error(), INTERNAL)
}
