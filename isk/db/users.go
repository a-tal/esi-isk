package db

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
)

// UserCharacter describes a mapping between a user and a character
type UserCharacter struct {
	RefreshToken  string    `db:"refresh_token"`
	AccessToken   string    `db:"access_token"`
	OwnerHash     string    `db:"owner_hash"`
	CharacterID   int32     `db:"character_id"`
	AccessExpires time.Time `db:"access_expires"`
	LastProcessed time.Time `db:"last_processed"`
}

// SaveUserCharacter attempts to save the UserCharacter in the db
func SaveUserCharacter(ctx context.Context, user *UserCharacter) error {
	prevChar, err := getUserCharacter(ctx, user.CharacterID)
	if err != nil {
		// new user
		return saveNewUserCharacter(ctx, user)
	}

	if prevChar != nil && user.OwnerHash == prevChar.OwnerHash {
		return updateUserCharacter(ctx, user)
	}

	if err := deleteUserCharacter(ctx, user.CharacterID); err != nil {
		log.Printf("failed to delete previous user: %+v", err)
		return err
	}

	return saveNewUserCharacter(ctx, user)
}

func updateUserCharacter(ctx context.Context, user *UserCharacter) error {
	return executeNamed(
		ctx,
		cx.StmtUpdateUser,
		map[string]interface{}{
			"character_id":   user.CharacterID,
			"refresh_token":  user.RefreshToken,
			"access_token":   user.AccessToken,
			"access_expires": user.AccessExpires,
			"owner_hash":     user.OwnerHash,
		},
	)
}

// save the newly created (or replaced) user
func saveNewUserCharacter(ctx context.Context, user *UserCharacter) error {
	return executeNamed(
		ctx,
		cx.StmtCreateUser,
		map[string]interface{}{
			"refresh_token":  user.RefreshToken,
			"access_token":   user.AccessToken,
			"access_expires": user.AccessExpires,
		},
	)
}

// pull the known user for this characterID
func getUserCharacter(
	ctx context.Context,
	characterID int32,
) (*UserCharacter, error) {
	res, err := executeNamedResult(
		ctx,
		cx.StmtGetUser,
		map[string]interface{}{"character_id": characterID},
	)

	if err != nil {
		return nil, err
	}

	for res.Next() {
		user := &UserCharacter{}
		if err := res.StructScan(user); err != nil {
			return nil, err
		}
		return user, nil
	}

	return nil, errors.New("User not found")
}

func deleteUserCharacter(ctx context.Context, charID int32) error {
	return executeNamed(
		ctx,
		cx.StmtDeleteUser,
		map[string]interface{}{"character_id": charID},
	)
}
