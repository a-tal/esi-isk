package worker

import (
	"context"
	"net/http"

	"github.com/antihax/goesi"
	"github.com/gregjones/httpcache"

	"github.com/a-tal/esi-isk/isk/cx"
)

// NewClient creates a new (cached) http client for ESI
func NewClient(ctx context.Context) *goesi.APIClient {
	cache := ctx.Value(cx.Cache).(httpcache.Cache)
	opts := ctx.Value(cx.Opts).(*cx.Options)

	transport := httpcache.NewTransport(cache)
	httpClient := &http.Client{Transport: transport}

	client := goesi.NewAPIClient(
		httpClient,
		// XXX version lookup here
		"esi-isk/0.0.1 <https://github.com/a-tal/esi-isk/>",
	)
	client.ChangeBasePath(opts.ESI)

	return client
}

/* XXX: WORKER: check owner hash in worker job, drop any mismatches */
