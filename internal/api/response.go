package api

import (
	"encoding/json"
	"net/http"
)

// EncodeJSONResponse encodes the given interface as a JSON response with the specified status and writes it to the provided http.ResponseWriter.
//
// Parameters:
// - i: the interface to be encoded as JSON
// - status: the HTTP status code to be set in the response header
// - w: the http.ResponseWriter to write the encoded JSON response to
func EncodeJSONResponse(i interface{}, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(i)
}
