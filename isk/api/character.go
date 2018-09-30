package api

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/a-tal/esi-isk/isk/db"
)

// CharacterDetails returns JSON describing the character
func CharacterDetails(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		charID, err := getCharID(r)
		if err != nil || charID < 1 {
			write400(w)
			return
		}

		c, err := db.GetCharDetails(ctx, charID)
		if err != nil {
			log.Printf("failed to get character details: %+v", err)
			write500(w)
			return
		}

		p, err := db.GetPreferences(ctx, "d", charID)
		if err == nil {
			if pErr := checkPassphrase(r, c, p); pErr != nil {
				write403(w)
				return
			}
		}

		writeJSON(ctx, w, c)
	}
}

// getCharID reads the "c" query arg
func getCharID(r *http.Request) (int32, error) {
	charID, err := strconv.ParseInt(r.URL.Query().Get("c"), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(charID), nil
}
