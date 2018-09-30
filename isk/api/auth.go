package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	oidc "github.com/coreos/go-oidc"
	sessions "github.com/goincremental/negroni-sessions"
	"github.com/twinj/uuid"
	"golang.org/x/oauth2"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
)

// StateStore stores state uuids we've given out
type StateStore struct {
	lock   *sync.Mutex
	states map[string]time.Time
}

// NewStateStore returns a new StateStore
func NewStateStore() *StateStore {
	ss := &StateStore{
		lock:   &sync.Mutex{},
		states: map[string]time.Time{},
	}
	go ss.maintenance()
	return ss
}

// NewProvider creates our oidc provider
func NewProvider(ctx context.Context) context.Context {
	return ctx // HACK: remove ones sso-issues#41 is done

	provider, err := oidc.NewProvider(ctx, "http://login.eveonline.com")
	if err != nil {
		log.Fatalf("failed to create provider: %+v", err)
	}

	ctx = context.WithValue(ctx, cx.Provider, provider)
	return ctx
}

// maintenance ensures we don't leak memory storing state uuids forever
func (s *StateStore) maintenance() {
	for {
		time.Sleep(30 * time.Second)
		s.prune()
	}
}

// prune will remove any stale states from memory
func (s *StateStore) prune() {
	s.lock.Lock()
	defer s.lock.Unlock()

	cutoff := stateCutoff()
	toPrune := []string{}
	for state, ts := range s.states {
		if ts.Before(cutoff) {
			toPrune = append(toPrune, state)
		}
	}

	for _, p := range toPrune {
		log.Printf("pruning old state: %s ts: %s", p, s.states[p])
		delete(s.states, p)
	}
}

func knownState(ctx context.Context, state string) bool {
	ss := ctx.Value(cx.StateStore).(*StateStore)
	ss.lock.Lock()
	defer ss.lock.Unlock()

	ts, found := ss.states[state]
	if !found {
		return false
	}

	delete(ss.states, state)

	return ts.After(stateCutoff())
}

func stateCutoff() time.Time {
	return time.Now().UTC().Add(-time.Duration(300) * time.Second)
}

func newState(ctx context.Context) string {
	state := uuid.NewV4().String()
	ss := ctx.Value(cx.StateStore).(*StateStore)
	ss.lock.Lock()
	ss.states[state] = time.Now().UTC()
	ss.lock.Unlock()
	return state
}

// NewLogin creates a new state and throws the user into the oauth flow
func NewLogin(ctx context.Context) http.HandlerFunc {
	opts := ctx.Value(cx.Opts).(*cx.Options)
	return func(w http.ResponseWriter, r *http.Request) {
		if opts.Auth == nil {
			write(w, 500, []byte("auth is not configured"))
			return
		}

		url := opts.Auth.AuthCodeURL(newState(ctx), oauth2.AccessTypeOffline)
		http.Redirect(w, r.WithContext(ctx), url, 302)
	}
}

// Callback receives oauth returns from sso
func Callback(ctx context.Context) http.HandlerFunc {
	opts := ctx.Value(cx.Opts).(*cx.Options)
	return func(w http.ResponseWriter, r *http.Request) {

		if opts.Auth == nil {
			write(w, 500, []byte("auth is not configured"))
			return
		}

		state := r.FormValue("state")
		code := r.FormValue("code")

		if !knownState(ctx, state) {
			write(w, 400, []byte("invalid state"))
			return
		}

		tok, err := opts.Auth.Exchange(ctx, code)
		if err != nil {
			write(w, 500, []byte("failed to complete token exchange"))
			return
		}

		user, err := userFromToken(ctx, tok)
		if err != nil {
			write(w, 500, []byte("failed to create new user"))
			return
		}

		if err := db.SaveUser(ctx, user); err != nil {
			write(w, 500, []byte("failed to save new user"))
			return
		}

		session := sessions.GetSession(r)
		session.Set("c", user.CharacterID)

		http.Redirect(w, r.WithContext(ctx), fmt.Sprintf(
			"/#prefs&c=%d&t=d",
			user.CharacterID,
		), 302)
	}
}

// userFromToken creates a userCharacter from the oauth2.Token
func userFromToken(
	ctx context.Context,
	t *oauth2.Token,
) (*db.User, error) {
	charID, owner, err := getOwnerFromJWT(ctx, t)
	if err != nil {
		return nil, err
	}

	user := &db.User{
		CharacterID:   charID,
		OwnerHash:     owner,
		RefreshToken:  t.RefreshToken,
		AccessToken:   t.AccessToken,
		AccessExpires: t.Expiry,
	}

	return user, nil
}

// getOwnerFromJWT parses the JWT for the character ID and owner hash
func getOwnerFromJWT(ctx context.Context, t *oauth2.Token) (
	charID int32,
	owner string,
	err error,
) {
	// HACK: remove once ccpgames/sso-issues#41 is done
	return VerifyTokenHack(ctx, t.AccessToken)

	verifier := ctx.Value(cx.Verifier).(*oidc.IDTokenVerifier)

	idToken, err := verifier.Verify(ctx, t.AccessToken)
	if err != nil {
		log.Printf("failed to verify token: %+v", err)
		return charID, owner, err
	}

	// Extract custom claims
	var claims struct {
		Scopes    []string `json:"scp"`
		Subject   string   `json:"sub"`
		OwnerHash string   `json:"owner"`
	}
	// TODO: verify scopes
	if claimErr := idToken.Claims(&claims); claimErr != nil {
		log.Printf("failed to parse claims from JWT: %+v", claimErr)
		return charID, owner, err
	}

	char, err := parseCharacterID(claims.Subject)
	if err != nil {
		log.Printf("failed to parse characterID from claim data: %+v", err)
		return char, owner, err
	}

	return char, claims.OwnerHash, nil
}

func parseCharacterID(sub string) (int32, error) {
	subSplit := strings.Split(sub, ":")
	if len(subSplit) != 3 {
		return 0, errors.New("sub claim is malformed")
	}

	charID, err := strconv.ParseInt(subSplit[2], 10, 32)
	return int32(charID), err
}

// VerifyModel is needed as a hack until an sso issue is fixed
type VerifyModel struct {
	CharacterID        int32
	CharacterOwnerHash string
}

// VerifyTokenHack pulls the charID and owner hash from the JWT
func VerifyTokenHack(ctx context.Context, token string) (
	charID int32,
	owner string,
	err error,
) {
	client := ctx.Value(cx.SSOClient).(*http.Client)
	res, err := client.Get(
		fmt.Sprintf("https://esi.evetech.net/verify/?token=%s", token),
	)
	if err != nil {
		log.Printf("failed to verify token (HACK): %+v", err)
		return 0, "", err
	}

	model := &VerifyModel{}
	dec := json.NewDecoder(res.Body)

	if err := dec.Decode(model); err != nil {
		log.Printf("failed to JSON decode verify response: %+v", err)
		return 0, "", err
	}

	if err := res.Body.Close(); err != nil {
		log.Printf("failed to close response body: %+v", err)
	}

	return model.CharacterID, model.CharacterOwnerHash, nil
}
