package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func write(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Printf("failed to write %d response: %+v", status, err)
	}
}

func write500(w http.ResponseWriter) {
	write(w, 500, []byte("internal error"))
}

func write400(w http.ResponseWriter) {
	write(w, 400, []byte("request error"))
}

func writeJSON(w http.ResponseWriter, res interface{}) {
	asJSON, err := json.Marshal(res)
	if err != nil {
		write500(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	write(w, 200, asJSON)
}
