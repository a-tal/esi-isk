package db

import (
	"context"
	"sort"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
)

// Donation describes a one time ISK transfer
type Donation struct {
	// ID is the transaction ID
	ID int64 `db:"transaction_id" json:"id"`

	// Donator who sent the recipient this isk
	Donator int32 `db:"donator" json:"donator"`

	// Recipient is who received the donation
	Recipient int32 `db:"receiver" json:"receiver"`

	// Timestamp of when this tranfer occurred
	Timestamp time.Time `db:"timestamp" json:"timestamp"`

	// Note or memo that came with the transfer
	Note string `db:"note" json:"note,omitempty"`

	// Amount of ISK transferred
	Amount float64 `db:"amount" json:"amount"`
}

// Donations are time sorted
type Donations []*Donation

func (d Donations) Len() int      { return len(d) }
func (d Donations) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d Donations) Less(i, j int) bool {
	return d[i].Timestamp.After(d[j].Timestamp)
}

// GetCharDonations returns donations FOR the character
func GetCharDonations(ctx context.Context, charID int32) (Donations, error) {
	return getDonations(ctx, charID, cx.StmtCharDonations)
}

// GetCharDonated returns donations FROM the character
func GetCharDonated(ctx context.Context, charID int32) (Donations, error) {
	return getDonations(ctx, charID, cx.StmtCharDonated)
}

// GetCharStandingISK returns the amount donated towards improving standing
func GetCharStandingISK(ctx context.Context, charID int32) (float64, error) {
	donations, err := getDonations(ctx, charID, cx.StmtCharStandingISK)
	if err != nil {
		return 0, err
	}

	total := float64(0)
	for _, d := range donations {
		total += d.Amount
	}
	return total, nil
}

// GetStaleDonations returns donations from more than 30 days ago
func GetStaleDonations(ctx context.Context) (Donations, error) {
	rows, err := queryNamedResult(ctx, cx.StmtGetStaleDonations, nil)
	if err != nil {
		return nil, err
	}

	res, err := scan(rows, func() interface{} { return &Donation{} })
	if err != nil {
		return nil, err
	}
	donations := Donations{}
	for _, i := range res {
		donations = append(donations, i.(*Donation))
	}

	return donations, nil
}

func getDonations(ctx context.Context, charID int32, key cx.Key) (
	Donations,
	error,
) {
	rows, err := queryNamedResult(ctx, key, map[string]interface{}{
		"character_id": charID,
	})

	if err != nil {
		return nil, err
	}

	res, err := scan(rows, func() interface{} { return &Donation{} })
	if err != nil {
		return nil, err
	}
	donations := Donations{}
	for _, i := range res {
		d := i.(*Donation)
		d.Amount = round2(d.Amount)
		donations = append(donations, d)
	}

	sort.Sort(donations)
	return donations, nil
}

// SaveDonation stores a donation in the database
func SaveDonation(ctx context.Context, donation *Donation) error {
	return executeNamed(ctx, cx.StmtAddDonation, map[string]interface{}{
		"transaction_id": donation.ID,
		"donator":        donation.Donator,
		"receiver":       donation.Recipient,
		"timestamp":      donation.Timestamp,
		"note":           donation.Note,
		"amount":         donation.Amount,
	})
}

// PruneDonation removes a donation by ID
func PruneDonation(ctx context.Context, donation *Donation) error {
	return executeNamed(ctx, cx.StmtRemoveDonation, map[string]interface{}{
		"transaction_id": donation.ID,
	})
}
