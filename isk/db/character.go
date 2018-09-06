package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Affiliation links a character with a corporation and maybe alliance
type Affiliation struct {
	Character   *Name
	Corporation *Name
	Alliance    *Name
}

// Character describes the output format of known characters
type Character struct {
	// ID is the characterID of this donator/recipient
	ID int32 `json:"id"`

	// Name is the last checked name of the character
	Name string `json:"name,omitempty"`

	// CorporationID is the last checked corporation ID of the character
	CorporationID int32 `json:"corporation,omitempty"`

	// CorporationName is the last checked name of the corporation
	CorporationName string `json:"corporation_name,omitempty"`

	// AllianceID is the last checked alliance ID of the character
	AllianceID int32 `json:"alliance,omitempty"`

	// AllianceName is the last checked name of the alliance
	AllianceName string `json:"alliance_name,omitempty"`

	// Received donations and/or contracts
	Received int64 `json:"received,omitempty"`

	// ReceivedISK value of all donations plus contracts
	ReceivedISK float64 `json:"received_isk,omitempty"`

	// Donated is the number of times this character has donated to someone else
	Donated int64 `json:"donated,omitempty"`

	// DonatedISK is the value of all ISK donated
	DonatedISK float64 `json:"donated_isk,omitempty"`

	// LastDonated timestamp
	LastDonated time.Time `json:"last_donated,omitempty"`

	// LastReceived timestamp
	LastReceived time.Time `json:"last_received,omitempty"`
}

// CharacterRow describes Character as stored in the characters table
type CharacterRow struct {
	// ID is the characterID of this donator/recipient
	ID int32 `db:"character_id"`

	// CorporationID is the last checked corporation ID of the character
	CorporationID int32 `db:"corporation_id"`

	// AllianceID is the last checked alliance ID of the character
	AllianceID int32 `db:"alliance_id"`

	// Received donations and/or contracts
	Received int64 `db:"received"`

	// ReceivedISK value of all donations plus contracts
	ReceivedISK float64 `db:"received_isk"`

	// Donated is the number of times this character has donated to someone else
	Donated int64 `db:"donated"`

	// DonatedISK is the value of all ISK donated
	DonatedISK float64 `db:"donated_isk"`

	// LastDonated timestamp
	LastDonated pq.NullTime `db:"last_donated"`

	// LastReceived timestamp
	LastReceived pq.NullTime `db:"last_received"`
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
}

// SaveCharacters updates all totals in the characters table
func SaveCharacters(
	ctx context.Context,
	donations []*Donation,
	affiliations []*Affiliation,
) error {
	newCharacters := []*CharacterRow{}
	updatedCharacters := []*CharacterRow{}
	allCharacters := []int32{}

	for _, donation := range donations {
		for _, charID := range []int32{donation.Donator, donation.Recipient} {
			if inInt32(charID, allCharacters) {
				continue
			}
			allCharacters = append(allCharacters, charID)

			char, new := bindAffiliation(ctx, charID, affiliations)
			if new {
				newCharacters = append(newCharacters, char)
			} else {
				updatedCharacters = append(updatedCharacters, char)
			}
		}

		addToTotals(donation, newCharacters, updatedCharacters)
	}

	return saveCharacters(ctx, newCharacters, updatedCharacters)
}

func saveCharacters(
	ctx context.Context,
	newCharacters, updatedCharacters []*CharacterRow,
) error {
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

func bindAffiliation(
	ctx context.Context,
	charID int32,
	affiliations []*Affiliation,
) (row *CharacterRow, new bool) {
	aff := getAffiliation(charID, affiliations)

	char, err := getCharDetails(ctx, charID)

	if err != nil {
		new = true
		row = &CharacterRow{ID: charID}
	} else {
		row = char.toRow()
	}

	row.CorporationID = aff.Corporation.ID
	if aff.Alliance != nil {
		row.AllianceID = aff.Alliance.ID
	}

	return row, new
}

// addToTotals adds donation/received totals
func addToTotals(donation *Donation, characters ...[]*CharacterRow) {
	for _, chars := range characters {
		for _, char := range chars {
			if char.ID == donation.Donator {
				char.DonatedISK += donation.Amount
				char.Donated++
				if !char.LastDonated.Valid || char.LastDonated.Time.Before(
					donation.Timestamp) {
					char.LastDonated = pq.NullTime{Time: donation.Timestamp, Valid: true}
					char.LastDonated.Valid = true
				}
			} else if char.ID == donation.Recipient {
				char.ReceivedISK += donation.Amount
				char.Received++
				if !char.LastReceived.Valid || char.LastReceived.Time.Before(
					donation.Timestamp) {
					char.LastReceived = pq.NullTime{Time: donation.Timestamp, Valid: true}
					char.LastReceived.Valid = true
				}
			}
		}
	}

}

// newCharacter adds a new character to the characters table
func newCharacter(ctx context.Context, char *CharacterRow) error {
	return executeChar(ctx, char, cx.StmtCreateCharacter)
}

// updateCharacter updates a character in the characters table
func updateCharacter(ctx context.Context, char *CharacterRow) error {
	return executeChar(ctx, char, cx.StmtUpdateCharacter)
}

// executeChar is a DRY helper to create or update a character
func executeChar(ctx context.Context, char *CharacterRow, key cx.Key) error {
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

func getCharDetails(ctx context.Context, charID int32) (*Character, error) {
	rows, err := queryNamedResult(
		ctx,
		cx.StmtCharDetails,
		map[string]interface{}{"character_id": charID},
	)

	if err != nil {
		return nil, err
	}

	charRow, err := scanCharacterRow(rows)
	if err != nil {
		return nil, err
	}

	char, err := getCharacterNames(ctx, charRow)
	if err != nil {
		return nil, err
	}

	return char, nil
}

func scanCharacterRow(rows *sqlx.Rows) (*CharacterRow, error) {
	res, err := scan(rows, func() interface{} { return &CharacterRow{} })
	if err != nil {
		return nil, err
	}
	for _, i := range res {
		return i.(*CharacterRow), nil
	}

	return nil, errors.New("character not found")
}

// getCharacterNames fills in the character, corporation and alliance names
func getCharacterNames(
	ctx context.Context,
	row *CharacterRow,
) (*Character, error) {
	ids := []int32{}
	for _, id := range []int32{row.ID, row.CorporationID, row.AllianceID} {
		if id > 0 {
			ids = append(ids, id)
		}
	}

	names, err := GetNames(ctx, ids...)
	if err != nil {
		return nil, err
	}

	char := row.toCharacter()

	for id, name := range names {
		if id == char.ID {
			char.Name = name
		} else if id == char.CorporationID {
			char.CorporationName = name
		} else if id == char.AllianceID {
			char.AllianceName = name
		} else {
			log.Printf("pulled unknown ID: %d, name: %s", id, name)
		}
	}

	return char, nil
}

func (c *CharacterRow) toCharacter() *Character {
	char := &Character{
		ID:            c.ID,
		CorporationID: c.CorporationID,
		AllianceID:    c.AllianceID,
		Received:      c.Received,
		ReceivedISK:   c.ReceivedISK,
		Donated:       c.Donated,
		DonatedISK:    c.DonatedISK,
	}
	if c.LastDonated.Valid {
		char.LastDonated = c.LastDonated.Time
	}
	if c.LastReceived.Valid {
		char.LastReceived = c.LastReceived.Time
	}
	return char
}

func (c *Character) toRow() *CharacterRow {
	return &CharacterRow{
		ID:            c.ID,
		CorporationID: c.CorporationID,
		AllianceID:    c.AllianceID,
		Received:      c.Received,
		ReceivedISK:   c.ReceivedISK,
		Donated:       c.Donated,
		DonatedISK:    c.DonatedISK,
		LastDonated: pq.NullTime{
			Time:  c.LastDonated,
			Valid: !c.LastDonated.IsZero(),
		},
		LastReceived: pq.NullTime{
			Time:  c.LastReceived,
			Valid: !c.LastReceived.IsZero(),
		},
	}
}
