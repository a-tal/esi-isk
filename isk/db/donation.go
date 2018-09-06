package db

import (
	"context"
	"log"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/jmoiron/sqlx"
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

func getCharDonations(ctx context.Context, charID int32) ([]*Donation, error) {
	return getDonations(ctx, charID, cx.StmtCharDonations)
}

func getCharDonated(ctx context.Context, charID int32) ([]*Donation, error) {
	return getDonations(ctx, charID, cx.StmtCharDonated)
}

func getDonations(ctx context.Context, charID int32, key cx.Key) (
	[]*Donation,
	error,
) {
	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)
	r, err := statements[key].Queryx(map[string]interface{}{
		"character_id": charID,
	})
	if err != nil {
		return nil, err
	}

	donations := []*Donation{}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("failed to close results: %+v", err)
		}
	}()

	for r.Next() {
		donation := &Donation{}
		if err := r.StructScan(donation); err != nil {
			return nil, err
		}
		donations = append(donations, donation)
	}

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
