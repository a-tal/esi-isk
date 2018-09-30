package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
)

// RFC1123 to be used with UTC timezone *only*
const RFC1123 = "Mon, 02 Jan 2006 15:04:05 GMT"

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

func write403(w http.ResponseWriter) {
	write(w, 403, []byte("request denied"))
}

func write405(w http.ResponseWriter) {
	write(w, 405, []byte("method not allowed"))
}

func writeJSON(ctx context.Context, w http.ResponseWriter, res interface{}) {
	asJSON, err := json.Marshal(res)
	if err != nil {
		write500(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeCacheHeaders(ctx, w)
	write(w, 200, asJSON)
}

func writeCacheHeaders(ctx context.Context, w http.ResponseWriter) {
	opts := ctx.Value(cx.Opts).(*cx.Options)
	now := time.Now().UTC()
	w.Header().Set("Last-Modified", now.Format(RFC1123))
	w.Header().Set(
		"Expires",
		now.Add(time.Duration(opts.CacheTime)*time.Second).Format(RFC1123),
	)
}
