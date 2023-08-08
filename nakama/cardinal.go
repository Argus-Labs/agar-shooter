package main

// cardinal.go wraps the http requests to some cardinal endpoints.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/argus-labs/world-engine/sign"
	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	createPersonaEndpoint     = "tx-create-persona"
	listTxEndpointsEndpoint   = "list/tx-endpoints"
	listReadEndpoints         = "list/read-endpoints"
	readPersonaSignerEndpoint = "read-persona-signer"

	readPersonaSignerStatusUnknown   = "unknown"
	readPersonaSignerStatusAvailable = "available"
	readPersonaSignerStatusAssigned  = "assigned"

	globalCardinalAddress string
	fakeSignerAddress     = "0xsigneraddress"
	fakePrivateKey        = "fake-private-key"

	ErrorPersonaSignerAvailable = errors.New("persona signer is available")
	ErrorPersonaSignerUnknown   = errors.New("persona signer is unknown.")
)

// Takes the Cardinal address variable and loads it into memory
func initCardinalAddress() error {
	globalCardinalAddress = os.Getenv(EnvCardinalAddr)
	if globalCardinalAddress == "" {
		return fmt.Errorf("must specify a cardinal server via %s", EnvCardinalAddr)
	}
	return nil
}

// Takes an endpoint name and appends it to the Cardinal address variable to produce a URL with which to access the endpoint
func makeURL(resource string) string {
	return fmt.Sprintf("%s/%s", globalCardinalAddress, resource)
}

func cardinalListEndpoints(path string) ([]string, error) {
	url := makeURL(path)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list endpoints (at %q) failed with status code %d: %v", url, resp.StatusCode, string(buf))
	}
	dec := json.NewDecoder(resp.Body)
	var endpoints []string
	if err := dec.Decode(&endpoints); err != nil {
		return nil, err
	}
	return endpoints, nil

}

// Produces a list of all Cardinal endpoints
func cardinalListAllEndpoints() ([]string, error) {
	var endpoints []string
	txs, err := cardinalListEndpoints(listTxEndpointsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Nakama: error fetching endpoint list:", err)
	}
	endpoints = append(endpoints, txs...)
	reads, err := cardinalListEndpoints(listReadEndpoints)
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, reads...)
	return endpoints, nil
}

// Creates a Persona for a given PersonaTag/name
func cardinalCreatePersona(ctx context.Context, nk runtime.NakamaModule, personaTag string) (tick uint64, err error) {
	createPersonaTx := struct {
		PersonaTag    string
		SignerAddress string
	}{
		PersonaTag:    personaTag,
		SignerAddress: fakeSignerAddress,
	}

	body, err := json.Marshal(createPersonaTx)
	if err != nil {
		fmt.Errorf("Error marshalling JSON for CreatePersonaTx: %v", err)
	}
	sp := &sign.SignedPayload{
		PersonaTag: personaTag,
		Namespace:  globalNamespace,
		Nonce:      1000,
		Signature:  "pk",
		Body:       body,
	}

	buf, err := sp.Marshal()
	if err != nil {
		return 0, fmt.Errorf("unable to marshal signed payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", makeURL(createPersonaEndpoint), bytes.NewReader(buf))
	if err != nil {
		return 0, fmt.Errorf("unable to make request to %q: %w", createPersonaEndpoint, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request to %q failed: %w", createPersonaEndpoint, err)
	} else if resp.StatusCode != 200 {
		buf, err := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("response is not 200: %v, %v", string(buf), err)
	}
	createPersonaResponse := struct {
		Status string
		Tick   uint64
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&createPersonaResponse); err != nil {
		return 0, fmt.Errorf("unable to decode response: %w", err)
	}
	if s := createPersonaResponse.Status; s != "ok" {
		return 0, fmt.Errorf("create persona failed with status %q", s)
	}
	return createPersonaResponse.Tick, nil
}

// IsUserIDSafeToQuery checks if the user ID is safe to query and updates the safety map accordingly.
func isUserIDSafeToQuery(userID string, joinTimeMap map[string]time.Time, isSafeToQueryMap map[string]bool) bool {
	isSafeToQuery, exists := isSafeToQueryMap[userID]

	// If the user ID is not in the safety map or is not safe, perform the calculation
	if !exists || !isSafeToQuery {
		// Perform the time calculation
		isSafe := joinTimeMap[userID].Add(time.Millisecond * 500).Before(time.Now())

		// Update the safety map
		isSafeToQueryMap[userID] = isSafe

		return isSafe
	}

	return true // Return true if the user ID was already marked as safe
}
