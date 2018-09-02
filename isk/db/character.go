package db

import (
	"time"
)

// Character describes someone who's sent or received ISK
type Character struct {
	// ID is the characterID of this donator/recipient
	ID int32 `db:"character_id" json:"id"`

	// Name is the last checked name of the character
	Name string `db:"character_name" json:"name"`

	// CorporationID is the last checked corporation ID of the character
	CorporationID int32 `db:"corporation_id" json:"corporation"`

	// CorporationName is the last checked corporation name of the character
	CorporationName string `db:"corporation_name" json:"corporation_name"`

	// Received donations and/or contracts
	Received int64 `db:"received" json:"received,omitempty"`

	// ReceivedISK value of all donations plus contracts
	ReceivedISK float64 `db:"received_isk" json:"received_isk,omitempty"`

	// Donated is the number of times this character has donated to someone else
	Donated int64 `db:"donated" json:"donated,omitempty"`

	// DonatedISK is the value of all ISK donated
	DonatedISK float64 `db:"donated_isk" json:"donated_isk,omitempty"`

	// Joined timestamp
	Joined *time.Time `db:"joined" json:"joined,omitempty"`

	// LastSeen timestamp
	LastSeen *time.Time `db:"last_seen" json:"last_seen,omitempty"`

	// LastDonated timestamp
	LastDonated *time.Time `db:"last_donated" json:"last_donated,omitempty"`

	// LastReceived timestamp
	LastReceived *time.Time `db:"last_received" json:"last_received,omitempty"`
}
