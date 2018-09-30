package worker

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/antihax/goesi"
	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"

	"github.com/a-tal/esi-isk/isk/api"
	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
)

// addClient adds an http client and goesi client to context
func addClient(ctx context.Context) context.Context {
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

	ctx = context.WithValue(ctx, cx.HTTPClient, httpClient)
	ctx = context.WithValue(ctx, cx.Client, client)
	return ctx
}

// Context adds the goesi client and auth to context
func Context(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, cx.DB, db.Connect(ctx))
	ctx = context.WithValue(ctx, cx.Statements, db.GetStatements(ctx))

	cache := httpcache.NewMemoryCache()
	ctx = context.WithValue(ctx, cx.Cache, cache)

	ctx = addClient(ctx)

	prices, err := newPrices(ctx)
	if err != nil {
		log.Fatalf("failed to fetch initial market prices: %+v", err)
	}
	ctx = context.WithValue(ctx, cx.Prices, prices)

	client := ctx.Value(cx.HTTPClient).(*http.Client)
	opts := ctx.Value(cx.Opts).(*cx.Options)

	ctx = context.WithValue(ctx, cx.Authenticator, goesi.NewSSOAuthenticatorV2(
		client,
		opts.Auth.ClientID,
		opts.Auth.ClientSecret,
		opts.Auth.RedirectURL,
		opts.Auth.Scopes,
	))

	return ctx
}

// Run -- main worker entry point -- this function does not return
func Run(ctx context.Context) {
	ctx = Context(ctx)

	loop := 0
	for {
		updateStandings(ctx, processUsers(ctx))
		time.Sleep(1 * time.Minute)
		loop++
		if loop%60 == 0 {
			pruneContracts(ctx)
			pruneDonations(ctx)
			loop = 0
		}
	}
}

func updateStandings(ctx context.Context, charIDs []int32) {
	opts := ctx.Value(cx.Opts).(*cx.Options)
	for _, charID := range charIDs {
		if charID == opts.CharacterID {
			continue
		}

		char, err := db.GetCharacter(ctx, charID)
		if err != nil {
			continue
		}

		standingISK, err := db.GetCharStandingISK(ctx, charID)
		if err != nil {
			continue
		}

		char.GoodStanding = standingISK > (char.ReceivedISK30 * float64(0.01))

		if err := db.SaveCharacter(ctx, char); err != nil {
			log.Printf("failed to save character %d: %+v", charID, err)
		}
	}
}

func processUsers(ctx context.Context) []int32 {
	processed := []int32{}
	users, err := db.GetUsersToProcess(ctx)
	if err != nil {
		log.Printf("could not pull users to process: %+v", err)
		return processed
	}

	for _, user := range users {
		// TODO: make this parallel

		ctx, err = addCharacterAuth(ctx, user)
		if err != nil {
			log.Printf("failed to get character auth: %+v", err)
			// delete the character? or track failures then delete
			continue
		}

		charIDs, err := pullCharacter(ctx, user)
		if err != nil {
			log.Printf("error pulling character %d: %+v", user.CharacterID, err)
		} else {
			for _, charID := range charIDs {
				isKnown := false
				for _, known := range processed {
					if known == charID {
						isKnown = true
					}
				}
				if !isKnown {
					processed = append(processed, charID)
				}
			}
		}
	}

	return processed
}

func getCharacterToken(
	ctx context.Context,
	user *db.User,
) (oauth2.TokenSource, error) {
	auth := ctx.Value(cx.Authenticator).(*goesi.SSOAuthenticator)

	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: user.RefreshToken,
		Expiry:       user.AccessExpires,
	}

	tokSrc := auth.TokenSource(token)
	tok, err := tokSrc.Token()
	if err != nil {
		return nil, err
	}

	charID, owner, err := api.VerifyTokenHack(ctx, tok.AccessToken)
	if err != nil {
		return nil, err
	}

	if charID != user.CharacterID || owner != user.OwnerHash {
		// TODO: delete these characters
		return nil, errors.New("characterID or owner hash mismatch")
	}

	return tokSrc, nil
}

func addCharacterAuth(
	ctx context.Context,
	user *db.User,
) (context.Context, error) {
	token, err := getCharacterToken(ctx, user)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, goesi.ContextOAuth2, token), nil
}

// pullCharacter is the top level function to pull a character's details
func pullCharacter(ctx context.Context, user *db.User) ([]int32, error) {
	log.Printf("pulling character: %d", user.CharacterID)

	charIDs, err := characterWallet(ctx, user)
	if err != nil {
		return charIDs, err
	}
	log.Printf("pulled character wallet: %d", user.CharacterID)

	contractCharIDs, err := characterContracts(ctx, user)
	if err != nil {
		return charIDs, err
	}
	log.Printf("pulled character contracts: %d", user.CharacterID)

	return append(charIDs, contractCharIDs...), db.SaveUser(ctx, user)
}
