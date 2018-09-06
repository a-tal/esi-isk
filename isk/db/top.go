package db

import (
	"context"
	"log"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/jmoiron/sqlx"
)

func getCharDetails(ctx context.Context, charID int32) (*Character, error) {
	char := &Character{}

	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)
	r, err := statements[cx.StmtCharDetails].Queryx(
		map[string]interface{}{"character_id": charID},
	)
	if err != nil {
		log.Printf("failed to pull character details: %+v", err)
		return nil, err
	}

	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("failed to close results: %+v", err)
		}
	}()

	for r.Next() {
		if err := r.StructScan(char); err != nil {
			log.Printf("failed to structscan character details: %+v", err)
			return nil, err
		}
		// break
		// break doesn't matter, there's only one here anyway...
		// linter complains /shrug
	}

	return getCharacterNames(ctx, char)
}

// getCharacterNames fills in the character, corporation and alliance names
func getCharacterNames(ctx context.Context, char *Character) (*Character, error) {
	log.Printf("pulling character names for character: %+v", char)

	ids := []int32{}
	for _, id := range []int32{char.ID, char.CorporationID, char.AllianceID} {
		if id > 0 {
			ids = append(ids, id)
		}
	}

	names, err := GetNames(ctx, ids...)
	if err != nil {
		log.Printf("could not pull names for %d, new character maybe", char.ID)
		return char, nil
	}

	for id, name := range names {
		if id == char.ID {
			char.Name = name
		} else if id == char.CorporationID {
			char.CorporationName = name
		} else if id == char.AllianceID {
			char.AllianceName = name
		}
	}

	return char, nil
}

// GetTopRecipients returns the top character IDs and isk values
func GetTopRecipients(ctx context.Context) ([]Character, error) {
	dbChars, err := queryCharISK(ctx, cx.StmtTopReceived)
	if err != nil {
		return nil, err
	}
	chars := []Character{}
	for _, char := range dbChars {
		chars = append(chars, Character{ID: char.id, ReceivedISK: char.isk})
	}
	return chars, nil
}

// GetTopDonators returns the top character IDs and isk values
func GetTopDonators(ctx context.Context) ([]Character, error) {
	chars, err := queryCharISK(ctx, cx.StmtTopDonated)
	if err != nil {
		return nil, err
	}
	characters := []Character{}
	for _, char := range chars {
		characters = append(characters, Character{ID: char.id, DonatedISK: char.isk})
	}
	return characters, nil
}

type charISK struct {
	id  int32
	isk float64
}

func queryCharISK(ctx context.Context, q cx.Key) ([]charISK, error) {
	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)

	res, err := statements[q].Queryx(map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			log.Printf("failed to close results: %+v", err)
		}
	}()

	return getCharISK(res), nil
}

func getCharISK(r *sqlx.Rows) []charISK {
	chars := []charISK{}
	for r.Next() {
		var charID int32
		var totalISK float64
		if err := r.Scan(&charID, &totalISK); err != nil {
			log.Printf("failed to scan getCharISK: %+v", err)
		} else {
			chars = append(chars, charISK{id: charID, isk: totalISK})
		}
	}
	return chars
}
