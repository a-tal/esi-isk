package db

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/a-tal/esi-isk/isk/cx"
)

// GetStatements prepares all queries for the global context
func GetStatements(ctx context.Context) map[cx.Key]*sqlx.NamedStmt {
	db := ctx.Value(cx.DB).(*sqlx.DB)
	statements := map[cx.Key]*sqlx.NamedStmt{}

	queries := map[cx.Key]string{
		cx.StmtTopReceived: `SELECT * FROM characters
ORDER BY received_isk DESC LIMIT 6`,

		cx.StmtTopDonated: `SELECT * FROM characters
ORDER BY donated_isk DESC LIMIT 6`,

		cx.StmtCharDetails: `SELECT * FROM characters
WHERE character_id = :character_id LIMIT 1`,

		// ISK IN
		cx.StmtCharDonations: `SELECT * FROM donations
WHERE receiver = :character_id`,
		cx.StmtCharContracts: `SELECT * FROM contracts
WHERE receiver = :character_id`,

		// ISK OUT
		cx.StmtCharDonated: `SELECT * FROM donations
WHERE donator = :character_id`,
		cx.StmtCharContracted: `SELECT * FROM contracts
WHERE donator = :character_id`,

		cx.StmtContractItems: `SELECT * FROM contractItems
WHERE contract_id = :contract_id LIMIT 1`,

		// USERS - user is a character w/ a token
		cx.StmtCreateUser: `INSERT INTO users (
    character_id,
    refresh_token,
    access_token,
    access_expires,
    owner_hash
) VALUES (
    :character_id,
    :refresh_token,
    :access_token,
    :access_expires,
    :owner_hash
)`,

		cx.StmtGetUser: `SELECT * FROM users
WHERE character_id = :character_id LIMIT 1`,

		cx.StmtGetUsers: `SELECT * FROM users
WHERE last_processed < NOW() - INTERVAL '1 hour' LIMIT 100`,

		cx.StmtGetNullUsers: `SELECT * FROM users
WHERE last_processed IS NULL LIMIT 100`,

		cx.StmtUpdateUser: `UPDATE users SET
    refresh_token = :refresh_token,
    access_token = :access_token,
    access_expires = :access_expires,
    character_id = :character_id,
    owner_hash = :owner_hash,
    last_journal_id = :last_journal_id,
    last_contract_id = :last_contract_id,
    last_processed = NOW()
WHERE character_id = :character_id`,

		cx.StmtDeleteUser: `DELETE FROM users WHERE character_id = :character_id`,

		cx.StmtAddDonation: `INSERT INTO donations (
    transaction_id,
    donator,
    receiver,
    "timestamp",
    note,
    amount
) VALUES (
    :transaction_id,
    :donator,
    :receiver,
    :timestamp,
    :note,
    :amount
)`,

		cx.StmtNewName: `INSERT INTO names (id, name) VALUES (:id, :name)`,

		cx.StmtUpdateName: `UPDATE names SET name = :name WHERE id = :id`,

		cx.StmtGetName: `SELECT * FROM names WHERE id = :id LIMIT 1`,

		cx.StmtCreateCharacter: `INSERT INTO characters (
    character_id,
    corporation_id,
    alliance_id,
    received,
    received_isk,
    donated,
    donated_isk,
    last_donated,
    last_received
) VALUES (
    :character_id,
    :corporation_id,
    :alliance_id,
    :received,
    :received_isk,
    :donated,
    :donated_isk,
    :last_donated,
    :last_received
)`,

		cx.StmtUpdateCharacter: `UPDATE characters SET
    corporation_id = :corporation_id,
    alliance_id = :alliance_id,
    received = :received,
    received_isk = :received_isk,
    donated = :donated,
    donated_isk = :donated_isk,
    last_donated = :last_donated,
    last_received = :last_received
WHERE character_id = :character_id`,

		cx.StmtAddContract: `INSERT INTO contracts (
    contract_id,
    donator,
    receiver,
    location,
    issued,
    expires,
    accepted,
    value,
    note
) VALUES (
    :contract_id,
    :donator,
    :receiver,
    :location,
    :issued,
    :expires,
    :accepted,
    :value,
    :note
)`,

		cx.StmtAddContractItems: `INSERT INTO contractItems (
    id,
    contract_id,
    type_id,
    item_id,
    quantity
) VALUES (
    :id,
    :contract_id,
    :type_id,
    :item_id,
    :quantity
)`,
	}

	for key, query := range queries {
		s, err := db.PrepareNamed(query)
		if err != nil {
			log.Fatalf("failed to prepare statement: %+v", err)
		}
		statements[key] = s
	}

	return statements
}
