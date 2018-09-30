package db

import (
	"context"
	"log"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/jmoiron/sqlx"
)

// GetTopRecipients returns the top character IDs and isk values
func GetTopRecipients(ctx context.Context) ([]*Character, error) {
	return getTop(
		ctx,
		cx.StmtTopReceived,
		func(c *CharacterRow) *Character {
			if c.ReceivedISK <= 0 {
				return nil
			}
			char := &Character{ID: c.ID, ReceivedISK: c.ReceivedISK}
			addValidTime(c, char)
			return char
		},
	)
}

// GetTopDonators returns the top character IDs and isk values
func GetTopDonators(ctx context.Context) ([]*Character, error) {
	return getTop(
		ctx,
		cx.StmtTopDonated,
		func(c *CharacterRow) *Character {
			if c.DonatedISK <= 0 {
				return nil
			}
			char := &Character{ID: c.ID, DonatedISK: c.DonatedISK}
			addValidTime(c, char)
			return char
		},
	)
}

func addValidTime(c *CharacterRow, char *Character) {
	if c.LastDonated.Valid {
		char.LastDonated = c.LastDonated.Time
	}
	if c.LastReceived.Valid {
		char.LastReceived = c.LastReceived.Time
	}
}

// getTop is a DRY helper for getting top donators and recipients
func getTop(
	ctx context.Context,
	key cx.Key,
	transform func(c *CharacterRow) *Character,
) ([]*Character, error) {
	chars, err := queryCharISK(ctx, key)
	if err != nil {
		return nil, err
	}

	characters := []*Character{}
	for _, charRow := range chars {
		char := transform(charRow)
		if char == nil {
			continue
		}
		name, err := GetName(ctx, char.ID)
		if err != nil {
			log.Printf("failed to lookup name for: %d", char.ID)
		} else {
			char.Name = name
		}
		characters = append(characters, char)
	}

	return characters, nil
}

func queryCharISK(ctx context.Context, q cx.Key) ([]*CharacterRow, error) {
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

	return scanCharacterRows(res)
}

func scanCharacterRows(rows *sqlx.Rows) ([]*CharacterRow, error) {
	res, err := scan(rows, func() interface{} { return &CharacterRow{} })
	if err != nil {
		return nil, err
	}
	chars := []*CharacterRow{}
	for _, i := range res {
		chars = append(chars, i.(*CharacterRow))
	}
	return chars, nil
}
