package worker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
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

func workerContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, cx.DB, db.Connect(ctx))
	ctx = context.WithValue(ctx, cx.Statements, db.GetStatements(ctx))

	cache := httpcache.NewMemoryCache()
	ctx = context.WithValue(ctx, cx.Cache, cache)

	ctx = addClient(ctx)

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
	ctx = workerContext(ctx)

	for {
		users, err := db.GetUsersToProcess(ctx)
		if err != nil {
			log.Printf("could not pull users to process: %+v", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, user := range users {
			// TODO: make this parallel

			ctx, err = addCharacterAuth(ctx, user)
			if err != nil {
				log.Printf("failed to get character auth: %+v", err)
				// delete the character? or track failures then delete
				continue
			}

			if err := pullCharacter(ctx, user); err != nil {
				log.Printf("error pulling character %d: %+v", user.CharacterID, err)
			}
		}

		time.Sleep(1 * time.Minute)
	}
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
func pullCharacter(ctx context.Context, user *db.User) error {
	log.Printf("pulling character: %d", user.CharacterID)

	if err := pullCharacterWallet(ctx, user); err != nil {
		return err
	}

	log.Printf("pulled character wallet: %d", user.CharacterID)

	if err := pullCharacterContracts(ctx, user); err != nil {
		return err
	}

	log.Printf("saving character: %d", user.CharacterID)
	err := db.SaveUser(ctx, user)

	log.Printf("saved character: %d. error: %+v", user.CharacterID, err)
	return err
}

func pullCharacterWallet(ctx context.Context, user *db.User) error {
	entries, err := getWalletJournal(ctx, user)
	if err != nil {
		return err
	}

	sort.Sort(entries)

	donations := parseForDonations(entries, user)

	setLastJournalID(entries, user)

	return saveWalletRun(ctx, donations, getNames(ctx, user, donations))
}

// getNames for all characters involved in the donations
func getNames(
	ctx context.Context,
	user *db.User,
	donations []*db.Donation,
) []*db.Affiliation {
	affiliations := []*db.Affiliation{}
	for _, donation := range donations {
		for _, charID := range []int32{donation.Donator, donation.Recipient} {
			isKnown := false
			for _, known := range affiliations {
				if known.Character.ID == charID || known.Corporation.ID == charID {
					isKnown = true
				}
			}
			if isKnown {
				continue
			}

			resolved, err := resolveNames(ctx, charID)
			if err != nil {
				log.Printf("failed to resolve names: %+v", err)
			} else {
				affiliations = append(affiliations, resolved)
			}
		}
	}
	return affiliations
}

// resolveNames gets the name of the charID,
// which might be a corp or allianceID. it will
// also resolve upwards, so corp+alliance in case
// of charID, and allianceID in case of corp
func resolveNames(ctx context.Context, charID int32) (*db.Affiliation, error) {

	aff := &db.Affiliation{}

	ret, err := resolveName(ctx, charID)
	if err != nil {
		return nil, err
	}

	for _, res := range ret {

		if res.Category == "corporation" {
			aff.Corporation = &db.Name{ID: res.Id, Name: res.Name}

			allianceID, allianceName := resolveCorporation(ctx, res.Id)
			if allianceID > 0 {
				aff.Alliance = &db.Name{ID: allianceID, Name: allianceName}
			}

		} else if res.Category == "character" {
			aff.Character = &db.Name{ID: res.Id, Name: res.Name}

			corpID, corpName := resolveCharacter(ctx, res.Id)
			if corpID > 0 { // 0 == error in lookup
				aff.Corporation = &db.Name{ID: corpID, Name: corpName}
				allianceID, allianceName := resolveCorporation(ctx, corpID)
				if allianceID > 0 { // 0 == error or not in alliance
					aff.Alliance = &db.Name{ID: allianceID, Name: allianceName}
				}
			}

		} else {
			// hopefully this doesn't happen
			log.Printf(
				"character received donation from %d who is a %s",
				charID,
				res.Category,
			)
		}
	}

	return aff, nil
}

// return the ID and name of the corporation's alliance
func resolveCorporation(ctx context.Context, corpID int32) (int32, string) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)

	ret, _, err := client.ESI.CorporationApi.GetCorporationsCorporationId(
		ctx,
		corpID,
		nil,
	)
	if err != nil {
		log.Printf("failed to get corporation by ID: %+v", err)
		return 0, ""
	}

	if ret.AllianceId > 0 {
		allianceNameRes, err := resolveName(ctx, ret.AllianceId)
		if err != nil {
			log.Printf("failed to resolve the alliance of corp: %d", corpID)
			return 0, ""
		}
		for _, res := range allianceNameRes {
			if res.Category == "alliance" && res.Id == ret.AllianceId {
				return ret.AllianceId, res.Name
			}
		}
	}

	return 0, ""
}

// resolveCharacter returns the ID and name of the character's corporation
func resolveCharacter(ctx context.Context, charID int32) (int32, string) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)

	ret, _, err := client.ESI.CharacterApi.GetCharactersCharacterId(
		ctx,
		charID,
		nil,
	)

	if err != nil {
		log.Printf("failed to lookup character %d by ID: %+v", charID, err)
		return 0, ""
	}

	corpRes, err := resolveName(ctx, ret.CorporationId)
	if err != nil {
		log.Printf("failed to resolve name of corp %d: %+v", ret.CorporationId, err)
		return 0, ""
	}

	for _, res := range corpRes {
		if res.Category == "corporation" {
			return res.Id, res.Name
		}
	}

	return 0, ""
}

func resolveName(ctx context.Context, charID ...int32) (
	[]esi.PostUniverseNames200Ok,
	error,
) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	ret, _, err := client.ESI.UniverseApi.PostUniverseNames(ctx, charID, nil)
	return ret, err
}

func getWalletJournal(
	ctx context.Context,
	user *db.User,
) (walletDonationEntries, error) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	entries, r, err := client.ESI.WalletApi.GetCharactersCharacterIdWalletJournal(
		ctx,
		user.CharacterID,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if !knownEntry(entries, user) {
		additional, err := expandWalletJournal(ctx, user, r)
		if err != nil {
			return nil, err
		}
		entries = append(entries, additional...)
	}

	return entries, nil
}

func getLastJournalID(user *db.User) (bool, int64) {
	return user.LastJournalID.Valid, user.LastJournalID.Int64
}

func parseForDonations(
	entries walletDonationEntries,
	user *db.User,
) []*db.Donation {
	donations := []*db.Donation{}
	hasLastID, lastID := getLastJournalID(user)
	for _, entry := range entries {
		if hasLastID && entry.Id == lastID {
			break
		}
		if entry.RefType == "player_donation" &&
			entry.SecondPartyId == user.CharacterID {
			donations = append(donations, &db.Donation{
				ID:        entry.Id,
				Donator:   entry.FirstPartyId,
				Recipient: user.CharacterID,
				Timestamp: entry.Date,
				Note:      entry.Reason,
				Amount:    entry.Amount,
			})
		}
	}
	return donations
}

func saveWalletRun(
	ctx context.Context,
	donations []*db.Donation,
	affiliations []*db.Affiliation,
) error {
	// NB: user is saved at a higher level

	for _, donation := range donations {
		if err := db.SaveDonation(ctx, donation); err != nil {
			return err
		}
	}

	if err := db.SaveNames(ctx, affiliations); err != nil {
		return err
	}

	return db.SaveCharacters(ctx, donations, affiliations)
}

type walletDonationEntries []esi.GetCharactersCharacterIdWalletJournal200Ok

func (w walletDonationEntries) Len() int      { return len(w) }
func (w walletDonationEntries) Swap(i, j int) { w[i], w[j] = w[j], w[i] }
func (w walletDonationEntries) Less(i, j int) bool {
	return w[i].Date.Before(w[j].Date)
}

func setLastJournalID(entries walletDonationEntries, user *db.User) {
	if len(entries) < 1 {
		return
	}
	if len(entries) >= 2 {
		log.Printf(
			"entry 0 ID: %d date: %s entry 1 ID: %d date: %s",
			entries[0].Id,
			entries[0].Date,
			entries[1].Id,
			entries[1].Date,
		)
	}
	user.LastJournalID = sql.NullInt64{
		Int64: entries[0].Id,
		Valid: true,
	}
}

// return true if we've seen one or more of these entries
func knownEntry(
	entries walletDonationEntries,
	user *db.User,
) bool {
	hasLastID, lastID := getLastJournalID(user)
	for _, entry := range entries {
		if hasLastID && entry.Id == lastID {
			return true
		}
	}
	return false
}

func expandWalletJournal(
	ctx context.Context,
	user *db.User,
	res *http.Response,
) (walletDonationEntries, error) {
	additional := walletDonationEntries{}
	xPagesRaw := res.Header.Get("X-Pages")
	if xPagesRaw == "" {
		return additional, nil
	}

	xPages64, err := strconv.ParseInt(xPagesRaw, 10, 32)
	if err != nil {
		return additional, err
	}
	xPages := int(xPages64)

	more := make(chan walletDonationEntries)
	errs := make(chan error)

	defer close(more)
	defer close(errs)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	started := 0
	for i := 2; i <= xPages; i++ {
		started++
		go func(i int) {
			entries, additionalErr := additionalWalletPage(ctx, user, int32(i))
			if additionalErr != nil {
				errs <- additionalErr
				return
			}
			more <- entries
		}(i)
	}

	for i := 0; i < started; i++ {
		select {
		case entries := <-more:
			additional = append(additional, entries...)
		case err = <-errs:
			cancel()
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return additional, nil
}

func additionalWalletPage(
	ctx context.Context,
	user *db.User,
	page int32,
) (walletDonationEntries, error) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	entries, _, err := client.ESI.WalletApi.GetCharactersCharacterIdWalletJournal(
		ctx,
		user.CharacterID,
		&esi.GetCharactersCharacterIdWalletJournalOpts{
			Page: optional.NewInt32(page),
		},
	)
	return entries, err
}

func pullCharacterContracts(ctx context.Context, user *db.User) error {
	// TODO
	return nil
}
