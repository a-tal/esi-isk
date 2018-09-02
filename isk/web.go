package isk

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/phyber/negroni-gzip/gzip"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"

	"github.com/a-tal/esi-isk/isk/api"
	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
)

func getAllowed(options *cx.Options) []string {
	proto := "http"
	if options.HTTPS {
		proto = "https"
	}

	return []string{
		fmt.Sprintf("%s://%s", proto, options.Hostname),
		fmt.Sprintf("%s://%s:%d", proto, options.Hostname, options.Port),
	}
}

// RunServer creates and runs the backend API server
func RunServer(ctx context.Context) {

	opts := ctx.Value(cx.Opts).(*cx.Options)

	ctx = context.WithValue(ctx, cx.DB, db.Connect(ctx))
	ctx = context.WithValue(ctx, cx.Statements, db.GetStatements(ctx))
	ctx = context.WithValue(ctx, cx.StateStore, api.NewStateStore())

	mux := http.NewServeMux()

	allowed := getAllowed(opts)

	mux.HandleFunc("/api/ping", api.Ping)
	mux.HandleFunc("/api/top", api.TopRecipients(ctx))
	mux.HandleFunc("/api/char", api.CharacterDetails(ctx))

	mux.HandleFunc("/signup", api.NewLogin(ctx))
	mux.HandleFunc("/callback", api.Callback(ctx))

	middleware := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),

		negroni.HandlerFunc(secure.New(secure.Options{
			FrameDeny:       true,
			AllowedHosts:    allowed,
			SSLRedirect:     opts.HTTPS,
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
			STSSeconds:      315360000,
			IsDevelopment:   opts.Debug,
			// ContentSecurityPolicy: "default-src 'self'",
		}).HandlerFuncWithNext),

		cors.New(cors.Options{
			AllowedOrigins:         allowed,
			AllowCredentials:       true,
			AllowOriginRequestFunc: nil,
			Debug:                  opts.Debug,
		}),

		gzip.Gzip(gzip.DefaultCompression),

		negroni.NewStatic(http.Dir("public")),
	)

	middleware.UseHandler(mux)

	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:              fmt.Sprintf(":%d", opts.Port),
			Handler:           middleware,
			ReadTimeout:       1 * time.Second,
			WriteTimeout:      1 * time.Second,
			ReadHeaderTimeout: 1 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
	}

	log.Fatal(server.ListenAndServe())
}
