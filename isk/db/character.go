package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/lib/pq"
)

// Affiliation links a character with a corporation and maybe alliance
type Affiliation struct {
	Character   *Name
	Corporation *Name
	Alliance    *Name
}

// Character describes someone who's sent or received ISK
type Character struct {
	// ID is the characterID of this donator/recipient
	ID int32 `db:"character_id" json:"id"`

	// Name is the last checked name of the character
	Name string `db:"-" json:"name"`

	// CorporationID is the last checked corporation ID of the character
	CorporationID int32 `db:"corporation_id" json:"corporation"`

	// CorporationName is the last checked name of the corporation
	CorporationName string `db:"-" json:"corporation_name"`

	// AllianceID is the last checked alliance ID of the character
	AllianceID int32 `db:"alliance_id" json:"alliance"`

	// AllianceName is the last checked name of the alliance
	AllianceName string `db:"-" json:"alliance_name"`

	// Received donations and/or contracts
	Received int64 `db:"received" json:"received,omitempty"`

	// ReceivedISK value of all donations plus contracts
	ReceivedISK float64 `db:"received_isk" json:"received_isk,omitempty"`

	// Donated is the number of times this character has donated to someone else
	Donated int64 `db:"donated" json:"donated,omitempty"`

	// DonatedISK is the value of all ISK donated
	DonatedISK float64 `db:"donated_isk" json:"donated_isk,omitempty"`

	// LastDonated timestamp
	LastDonated pq.NullTime `db:"last_donated" json:"last_donated,omitempty"`

	// LastReceived timestamp
	LastReceived pq.NullTime `db:"last_received" json:"last_received,omitempty"`
}

// CharDetails is the api return for a character
type CharDetails struct {
	Character *Character `json:"character"`

	// ISK IN
	Donations []*Donation `json:"donations,omitempty"`
	Contracts []*Contract `json:"contracts,omitempty"`

	// ISK OUT
	Donated    []*Donation `json:"donated,omitempty"`
	Contracted []*Contract `json:"contracted,omitempty"`
}

// GetCharDetails returns details for the character from pg
func GetCharDetails(ctx context.Context, charID int32) (*CharDetails, error) {
	// hmm
	char, err := getCharDetails(ctx, charID)
	if err != nil {
		return nil, err
	}

	contracts, err := getCharContracts(ctx, charID)
	if err != nil {
		return nil, err
	}

	contracted, err := getCharContracted(ctx, charID)
	if err != nil {
		return nil, err
	}

	donations, err := getCharDonations(ctx, charID)
	if err != nil {
		return nil, err
	}

	donated, err := getCharDonated(ctx, charID)
	if err != nil {
		return nil, err
	}

	details := &CharDetails{
		Character:  char,
		Donations:  donations,
		Donated:    donated,
		Contracts:  contracts,
		Contracted: contracted,
	}

	log.Printf("character: %+v", details.Character)

	return details, nil
}

func getAffiliation(charID int32, affiliations []*Affiliation) *Affiliation {
	for _, aff := range affiliations {
		if aff.Character.ID == charID {
			return aff
		}
	}
	// this should never happen
	panic(fmt.Errorf("no affiliation found for character %d", charID))
	return nil
}

// SaveCharacters updates all totals in the characters table
func SaveCharacters(
	ctx context.Context,
	donations []*Donation,
	affiliations []*Affiliation,
) error {
	// TODO: refactor this into smaller functions
	newCharacters := []*Character{}
	updatedCharacters := []*Character{}
	allCharacters := []int32{}

	for _, dono := range donations {
		// get affiliations, resolve names
		for _, charID := range []int32{dono.Donator, dono.Recipient} {
			known := false
			for _, knownChar := range allCharacters {
				if knownChar == charID {
					known = true
					break
				}
			}
			if known {
				continue
			}

			allCharacters = append(allCharacters, charID)

			aff := getAffiliation(charID, affiliations)

			char, err := getCharDetails(ctx, charID)

			newCharacter := false
			if err != nil {
				newCharacter = true
				char = &Character{ID: charID}
			}

			char.Name = aff.Character.Name
			char.CorporationID = aff.Corporation.ID
			char.CorporationName = aff.Corporation.Name

			if aff.Alliance != nil {
				char.AllianceID = aff.Alliance.ID
				char.AllianceName = aff.Alliance.Name
			}

			if newCharacter {
				newCharacters = append(newCharacters, char)
			} else {
				updatedCharacters = append(updatedCharacters, char)
			}

		}

		// add donation/received totals
		for _, characters := range [][]*Character{newCharacters, updatedCharacters} {
			for _, char := range characters {
				if char.ID == dono.Donator {
					char.DonatedISK += dono.Amount
					char.Donated++
					if !char.LastDonated.Valid || char.LastDonated.Time.Before(dono.Timestamp) {
						char.LastDonated = pq.NullTime{Time: dono.Timestamp, Valid: true}
						char.LastDonated.Valid = true
					}
				} else if char.ID == dono.Recipient {
					char.ReceivedISK += dono.Amount
					char.Received++
					if !char.LastReceived.Valid || char.LastReceived.Time.Before(dono.Timestamp) {
						char.LastReceived = pq.NullTime{Time: dono.Timestamp, Valid: true}
						char.LastReceived.Valid = true
					}
				}
			}
		}
	}

	failedChars := []string{}
	for _, char := range newCharacters {
		if err := newCharacter(ctx, char); err != nil {
			log.Printf("failed to save new character %d: %+v", char.ID, err)
			failedChars = append(failedChars, fmt.Sprintf("%d", char.ID))
		}
	}

	for _, char := range updatedCharacters {
		if err := updateCharacter(ctx, char); err != nil {
			log.Printf("failed to save updated character %d: %+v", char.ID, err)
			failedChars = append(failedChars, fmt.Sprintf("%d", char.ID))
		}
	}

	if len(failedChars) > 0 {
		return fmt.Errorf(
			"failed to save character(s): %s",
			strings.Join(failedChars, ", "),
		)
	}

	return nil
}

// newCharacter adds a new character to the characters table
func newCharacter(ctx context.Context, char *Character) error {
	return executeChar(ctx, char, cx.StmtCreateCharacter)
}

// updateCharacter updates a character in the characters table
func updateCharacter(ctx context.Context, char *Character) error {
	return executeChar(ctx, char, cx.StmtUpdateCharacter)
}

// executeChar is a DRY helper to create or update a character
func executeChar(ctx context.Context, char *Character, key cx.Key) error {
	return executeNamed(ctx, key, map[string]interface{}{
		"character_id":   char.ID,
		"corporation_id": char.CorporationID,
		"alliance_id":    char.AllianceID,
		"received":       char.Received,
		"received_isk":   char.ReceivedISK,
		"donated":        char.Donated,
		"donated_isk":    char.DonatedISK,
		"last_donated":   char.LastDonated,
		"last_received":  char.LastReceived,
	})
}
