package api

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/a-tal/esi-isk/isk/db"
)

// CharacterDetails returns JSON describing the character
// XXX add pagination! (donations, contracts, etc)
func CharacterDetails(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		char := r.URL.Query().Get("c")
		charID, err := strconv.ParseInt(char, 10, 32)
		if err != nil || charID < 1 {
			write400(w)
			return
		}

		// XXX cache this
		character, err := db.GetCharDetails(ctx, int32(charID))
		if err != nil {
			log.Printf("failed to get character details: %+v", err)
			write500(w)
			return
		}
		writeJSON(w, character)
	}
}
