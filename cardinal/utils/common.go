package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/rs/zerolog/log"
)

func Must(err error) {
	if err != nil {
		log.Fatal().Err(err)
	}
}

func WriteError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(500)
	payload := struct {
		Msg string
		Err string
	}{Msg: msg, Err: err.Error()}

	enc := json.NewEncoder(w)
	if err := enc.Encode(payload); err != nil {
		WriteError(w, "can't encode", err)
		return
	}

	log.Error().Err(err).Msg(msg)
}

func WriteResult(w http.ResponseWriter, v any) {
	if s, ok := v.(string); ok {
		v = struct{ Msg string }{Msg: s}
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		WriteError(w, "can't encode", err)
		return
	}
}

func Decode(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

func DecodeMsg[T any](r *http.Request, msg *T) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	// Try to decode the request body into struct T.
	// Errors if msg has unknown fields.
	err := decoder.Decode(&msg)
	if err != nil {
		return err
	}

	// Check that all fields are present
	fields := reflect.ValueOf(msg).Elem()
	for i := 0; i < fields.NumField(); i++ {
		if fields.Field(i).IsZero() {
			return errors.New("some msg field(s) are missing")
		}
	}

	return nil
}
