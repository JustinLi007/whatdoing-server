package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type Envelope map[string]any

func WriteJson(w http.ResponseWriter, statusCode int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	log.Printf("%v", string(js))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(js)
	return err
}
