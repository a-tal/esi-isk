package isk

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"gopkg.in/tylerb/graceful.v1"

	"github.com/a-tal/esi-isk/isk/api"
	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
	"github.com/a-tal/esi-isk/isk/worker"
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

func getCache(ctx context.Context) (*cache.Client, cache.Adapter) {
	opts := ctx.Value(cx.Opts).(*cx.Options)

	adapter, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(opts.CacheResp),
	)
	if err != nil {
		log.Fatal(err)
	}

	client, err := cache.NewClient(
		cache.ClientWithAdapter(adapter),
		cache.ClientWithTTL(time.Duration(opts.CacheTime)*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	return client, adapter
}

// RunServer creates and runs the backend API server
func RunServer(ctx context.Context) {

	opts := ctx.Value(cx.Opts).(*cx.Options)

	ctx = context.WithValue(ctx, cx.DB, db.Connect(ctx))
	ctx = context.WithValue(ctx, cx.Statements, db.GetStatements(ctx))
	ctx = context.WithValue(ctx, cx.StateStore, api.NewStateStore())

	if err := InitialSetup(ctx); err != nil {
		log.Fatalf("failed to initialize db: %+v", err)
	}

	mux := http.NewServeMux()

	respCache, adapter := getCache(ctx)
	ctx = context.WithValue(ctx, cx.Adapter, adapter)

	mux.HandleFunc("/api/ping", api.Ping)
	mux.Handle("/api/prefs", api.Preferences(ctx))
	mux.Handle("/api/top", respCache.Middleware(api.TopRecipients(ctx)))
	mux.Handle("/api/char", respCache.Middleware(api.CharacterDetails(ctx)))
	mux.Handle("/api/custom", respCache.Middleware(api.Custom(ctx)))

	mux.HandleFunc("/signup", api.NewLogin(ctx))
	mux.HandleFunc("/callback", api.Callback(ctx))

	middleware := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),

		negroni.HandlerFunc(secure.New(secure.Options{
			FrameDeny:       true,
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
			STSSeconds:      315360000,
			IsDevelopment:   opts.Debug,
			ContentSecurityPolicy: "default-src 'self' script-src 'unsafe-inline' " +
				"img-src 'self' imageserver.eveonline.com",
		}).HandlerFuncWithNext),

		cors.New(cors.Options{
			AllowedOrigins:         getAllowed(opts),
			AllowCredentials:       true,
			AllowOriginRequestFunc: nil,
			Debug:                  opts.Debug,
		}),

		gzip.Gzip(gzip.DefaultCompression),

		negroni.NewStatic(http.Dir("public")),
	)

	store := cookiestore.New([]byte(opts.AppSecret))
	middleware.Use(sessions.Sessions("esi-isk", store))

	middleware.UseHandler(mux)

	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:              fmt.Sprintf(":%d", opts.Port),
			Handler:           middleware,
			ReadTimeout:       1 * time.Second,
			WriteTimeout:      5 * time.Second,
			ReadHeaderTimeout: 1 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
	}

	log.Fatal(server.ListenAndServe())
}

// InitialSetup ensures the owning character exists in the db
func InitialSetup(ctx context.Context) error {
	opts := ctx.Value(cx.Opts).(*cx.Options)
	_, err := db.GetCharacter(ctx, opts.CharacterID)

	if err == nil {
		return nil
	}
	ctx = worker.Context(ctx)

	names, err := worker.ResolveName(ctx, opts.CharacterID)
	if err != nil {
		return err
	}

	var charName string
	for _, name := range names {
		if name.Category == "character" && name.Id == opts.CharacterID {
			charName = name.Name
			break
		}
	}

	if charName == "" {
		return fmt.Errorf(
			"could not lookup name for owning character ID: %d",
			opts.CharacterID,
		)
	}

	corp, corpName := worker.ResolveCharacter(ctx, opts.CharacterID)

	aff := &db.Affiliation{
		Character:   &db.Name{ID: opts.CharacterID, Name: charName},
		Corporation: &db.Name{ID: corp, Name: corpName},
	}

	alliance, allianceName := worker.ResolveCorporation(ctx, corp)
	if alliance > 0 {
		aff.Alliance = &db.Name{ID: alliance, Name: allianceName}
	}

	if err := db.SaveNames(ctx, []*db.Affiliation{aff}); err != nil {
		return err
	}

	log.Printf("creating owner character: %d", opts.CharacterID)
	return db.NewCharacter(ctx, &db.CharacterRow{
		ID:            opts.CharacterID,
		CorporationID: corp,
		AllianceID:    alliance,
		Received:      0,
		ReceivedISK:   0,
		Donated:       0,
		DonatedISK:    0,
		GoodStanding:  true,
	})
}
