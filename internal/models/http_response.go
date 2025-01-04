package models

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
}

// Marshalls the Type in Data and sends a response to the browser. This is the last
// step in the request/ response cycle.
func (hr *HttpResponse) Respond(w http.ResponseWriter) error {

	jsonData, err := json.Marshal(hr)

	if err != nil {
		return fmt.Errorf("error marshalling error %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err != nil {
		w.Write([]byte(fmt.Sprintf("error decoding recipe batch. %v", err)))
	}

	_, err = w.Write([]byte(jsonData))

	if err != nil {
		return fmt.Errorf("error marshalling error %v", err)
	}

	return nil

}
