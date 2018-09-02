package api

import (
	"log"
	"net/http"
)

// Ping returns a simple 200 ok
func Ping(w http.ResponseWriter, _req *http.Request) {
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Println("failed to write ping response")
	}
}
