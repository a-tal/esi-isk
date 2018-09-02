package db

import "time"

// Donation describes a one time ISK transfer
type Donation struct {
	// ID is the transaction ID
	ID int64 `db:"transaction_id" json:"id"`

	// Donator who sent the recipient this isk
	Donator int32 `db:"donator" json:"donator"`

	// Recipient is who received the donation
	Recipient int32 `db:"receiver" json:"receiver"`

	// Timestamp of when this tranfer occurred
	Timestamp *time.Time `db:"timestamp" json:"timestamp"`

	// Note or memo that came with the transfer
	Note string `db:"note" json:"note,omitempty"`

	// Amount of ISK transferred
	Amount float64 `db:"amount" json:"amount"`
}
