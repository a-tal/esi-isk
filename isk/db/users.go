package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/jmoiron/sqlx"
)

// User describes a mapping between a user and a character
type User struct {
	RefreshToken   string        `db:"refresh_token"`
	AccessToken    string        `db:"access_token"`
	OwnerHash      string        `db:"owner_hash"`
	CharacterID    int32         `db:"character_id"`
	LastJournalID  sql.NullInt64 `db:"last_journal_id"`
	LastContractID sql.NullInt64 `db:"last_contract_id"`
	AccessExpires  time.Time     `db:"access_expires"`
	LastProcessed  *time.Time    `db:"last_processed"`
}

// GetUsersToProcess returns all characters needing to be processed
func GetUsersToProcess(ctx context.Context) ([]*User, error) {
	users := []*User{}

	nullUsers, err := queryUsers(ctx, cx.StmtGetNullUsers)
	if err != nil {
		return nil, err
	}
	users = append(users, nullUsers...)

	updateUsers, err := queryUsers(ctx, cx.StmtGetUsers)
	if err != nil {
		return nil, err
	}
	users = append(users, updateUsers...)

	return users, nil
}

func queryUsers(
	ctx context.Context,
	key cx.Key,
) ([]*User, error) {
	rows, err := queryNamedResult(ctx, key, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	return scanUsers(rows)
}

// SaveUser attempts to save the User in the db
func SaveUser(ctx context.Context, user *User) error {
	prevChar, err := getUser(ctx, user.CharacterID)
	if err != nil {
		// new user
		log.Println("saving new character")
		return saveNewUser(ctx, user)
	}

	if prevChar != nil && user.OwnerHash == prevChar.OwnerHash {
		log.Println("updating known character")
		return updateUser(ctx, user)
	}

	if err := deleteUser(ctx, user.CharacterID); err != nil {
		log.Printf("failed to delete previous user: %+v", err)
		return err
	}

	log.Println("deleted old character, saving new character")
	return saveNewUser(ctx, user)
}

func updateUser(ctx context.Context, user *User) error {
	return executeNamed(ctx, cx.StmtUpdateUser, map[string]interface{}{
		"character_id":     user.CharacterID,
		"refresh_token":    user.RefreshToken,
		"access_token":     user.AccessToken,
		"access_expires":   user.AccessExpires,
		"owner_hash":       user.OwnerHash,
		"last_journal_id":  user.LastJournalID,
		"last_contract_id": user.LastContractID,
	})
}

// save the newly created (or replaced) user
func saveNewUser(ctx context.Context, user *User) error {
	return executeNamed(ctx, cx.StmtCreateUser, map[string]interface{}{
		"character_id":   user.CharacterID,
		"refresh_token":  user.RefreshToken,
		"access_token":   user.AccessToken,
		"access_expires": user.AccessExpires,
		"owner_hash":     user.OwnerHash,
	})
}

// pull the known user for this characterID
func getUser(
	ctx context.Context,
	characterID int32,
) (*User, error) {
	rows, err := queryNamedResult(
		ctx,
		cx.StmtGetUser,
		map[string]interface{}{"character_id": characterID},
	)

	if err != nil {
		return nil, err
	}

	users, err := scanUsers(rows)
	if err != nil {
		return nil, err
	} else if len(users) != 1 {
		return nil, errors.New("User not found")
	}

	return users[0], nil
}

func scanUsers(rows *sqlx.Rows) ([]*User, error) {
	res, err := scan(rows, func() interface{} { return &User{} })
	if err != nil {
		return nil, err
	}
	users := []*User{}
	for _, i := range res {
		users = append(users, i.(*User))
	}
	return users, nil
}

func deleteUser(ctx context.Context, charID int32) error {
	return executeNamed(
		ctx,
		cx.StmtDeleteUser,
		map[string]interface{}{"character_id": charID},
	)
}
