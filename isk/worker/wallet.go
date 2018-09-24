package worker

import (
	"context"
	"database/sql"
	"net/http"
	"sort"
	"strconv"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
)

func pullCharacterWallet(ctx context.Context, user *db.User) ([]int32, error) {
	charIDs := []int32{}

	entries, err := getWalletJournal(ctx, user)
	if err != nil {
		return charIDs, err
	}

	sort.Sort(entries)

	donations := parseForDonations(entries, user)

	if len(donations) > 0 {
		charIDs = append(charIDs, user.CharacterID)
	}

	for _, donation := range donations {
		charIDs = append(charIDs, donation.Donator)
	}

	setLastJournalID(entries, user)

	return charIDs, saveWalletRun(ctx, donations, getNames(ctx, user, donations))
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

	return db.SaveCharacterDonations(ctx, donations, affiliations)
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
