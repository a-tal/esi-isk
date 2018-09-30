package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"net/url"
	"sort"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
	sessions "github.com/goincremental/negroni-sessions"
	cache "github.com/victorspringer/http-cache"
)

// Preferences handles getting and setting user preferences
func Preferences(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			write405(w)
			return
		}

		session := sessions.GetSession(r)
		char := session.Get("c")
		if char == nil {
			if r.Method == http.MethodGet {
				http.Redirect(w, r, "/login", 302)
			} else {
				write403(w)
			}
			return
		}

		charID, ok := char.(int32)
		if !ok || charID < 1 {
			write403(w)
			return
		}

		if r.Method == http.MethodPost {
			updatePreferences(w, r.WithContext(ctx), charID)
		} else {
			writePreferences(w, r.WithContext(ctx), charID)
		}
	}
}

func writePreferences(w http.ResponseWriter, r *http.Request, charID int32) {
	p, err := getPreferences(w, r, charID)
	if err != nil {
		return
	}

	if p.Contracts != nil && p.Donations != nil {
		writeJSON(r.Context(), w, p)
	} else if p.Contracts != nil {
		writeJSON(r.Context(), w, p.Contracts)
	} else {
		writeJSON(r.Context(), w, p.Donations)
	}
}

func updatePreferences(w http.ResponseWriter, r *http.Request, charID int32) {
	t, err := getPrefType(r)
	if err != nil {
		write400(w)
		return
	}

	p, readErr := readPreferences(r, t)
	if readErr != nil {
		write400(w)
		return
	}

	ctx := r.Context()

	if err := db.SetPreferences(ctx, charID, p); err != nil {
		log.Printf("failed to set user preferences: %+v", err)
		write400(w)
	} else {
		dropCustomAPICache(ctx, charID, p, t)
		w.WriteHeader(204)
	}
}

func dropCustomAPICache(
	ctx context.Context,
	charID int32,
	p *db.Preferences,
	t string,
) {
	u := fmt.Sprintf("/api/custom?c=%d&t=%s", charID, t)
	dropCache(ctx, u)

	var passphrase string
	if t == "c" {
		passphrase = p.Contracts.Passphrase
	} else {
		passphrase = p.Donations.Passphrase
	}

	if passphrase != "" {
		dropCache(ctx, fmt.Sprintf("%s&p=%s", u, passphrase))
	}
}

func dropCache(ctx context.Context, path string) {
	u, err := url.Parse(path)
	if err == nil {
		adapter := ctx.Value(cx.Adapter).(cache.Adapter)
		sortURLParams(u)
		adapter.Release(generateKey(u.String()))
	}
}

func sortURLParams(URL *url.URL) {
	params := URL.Query()
	for _, param := range params {
		sort.Slice(param, func(i, j int) bool {
			return param[i] < param[j]
		})
	}
	URL.RawQuery = params.Encode()
}

func generateKey(URL string) uint64 {
	hash := fnv.New64a()
	if _, err := hash.Write([]byte(URL)); err != nil {
		log.Printf("failed to hash url: %+v", err)
	}
	return hash.Sum64()
}

func getPrefType(r *http.Request) (string, error) {
	t := r.URL.Query().Get("t")
	if t == "" {
		t = "d"
	} else if t != "d" && t != "c" && t != "a" {
		return "", errors.New("invalid preference type")
	}
	return t, nil
}

func readPreferences(r *http.Request, t string) (*db.Preferences, error) {
	if t == "a" {
		return readMultiplePrefs(r)
	}

	p, err := readSingularPrefs(r)
	if err != nil {
		return nil, err
	}

	if t == "d" {
		return &db.Preferences{Donations: p}, nil
	}
	return &db.Preferences{Contracts: p}, nil
}

func readSingularPrefs(r *http.Request) (*db.Prefs, error) {
	decoder := json.NewDecoder(r.Body)
	p := &db.Prefs{}
	if err := decoder.Decode(p); err != nil {
		return nil, err
	}

	if err := p.Sanity(r.Context()); err != nil {
		return nil, err
	}

	return p, nil
}

func readMultiplePrefs(r *http.Request) (*db.Preferences, error) {
	decoder := json.NewDecoder(r.Body)
	p := &db.Preferences{}
	if err := decoder.Decode(p); err != nil {
		return nil, err
	}

	if err := p.Donations.Sanity(r.Context()); err != nil {
		return nil, err
	}

	if err := p.Contracts.Sanity(r.Context()); err != nil {
		return nil, err
	}

	return p, nil
}

// getPreferences returns the db.Preferences for the charID or write an error
func getPreferences(w http.ResponseWriter, r *http.Request, charID int32) (
	*db.Preferences,
	error,
) {
	t, err := getPrefType(r)
	if err != nil {
		write400(w)
		return nil, err
	}

	prefs, err := db.GetPreferences(r.Context(), t, charID)
	if err != nil {
		if ue, ok := err.(*db.UserError); ok {
			write(w, ue.Code, ue.Msg)
			return nil, err
		}
		log.Printf("failed to get user preferences: %+v", err)
		write500(w)
		return nil, err
	}

	return prefs, nil
}
