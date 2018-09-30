package api

import (
	"context"
	"net/http"

	"github.com/a-tal/esi-isk/isk/db"
)

// TopRecipients returns JSON describing the current top donation recipients
func TopRecipients(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, _req *http.Request) {
		recipients, err := db.GetTopRecipients(ctx)
		if err != nil {
			write500(w)
			return
		}

		donators, err := db.GetTopDonators(ctx)
		if err != nil {
			write500(w)
			return
		}

		res := map[string][]*db.Character{
			"recipients": recipients,
			"donators":   donators,
		}

		writeJSON(ctx, w, res)
	}
}
