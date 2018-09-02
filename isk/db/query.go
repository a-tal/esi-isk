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
		cx.StmtTopReceived: `SELECT character_id, received_isk FROM characters
ORDER BY received_isk DESC LIMIT 6`,
		cx.StmtTopDonated: `SELECT character_id, donated_isk FROM characters
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
WHERE contract_id = :contract_id`,

		// USERS - user is a character w/ a token
		cx.StmtCreateUser: `INSERT INTO users (
    refresh_token,
    access_token,
    access_expires
) VALUES (
    :refresh_token,
    :access_token,
    :access_expires
)`,

		cx.StmtGetUser: `SELECT * FROM users WHERE character_id = :character_id`,

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
    last_processed = NOW()
WHERE character_id = :character_id`,

		cx.StmtDeleteUser: `DELETE FROM users WHERE character_id = :character_id`,
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

	log.Printf("character details: %+v", details)

	return details, nil
}

func getCharContracts(ctx context.Context, charID int32) ([]*Contract, error) {
	return getContracts(ctx, charID, cx.StmtCharContracts)
}

func getCharContracted(ctx context.Context, charID int32) ([]*Contract, error) {
	return getContracts(ctx, charID, cx.StmtCharContracted)
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

func getContracts(ctx context.Context, charID int32, key cx.Key) (
	[]*Contract,
	error,
) {
	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)
	r, err := statements[key].Queryx(map[string]interface{}{
		"character_id": charID,
	})
	if err != nil {
		return nil, err
	}

	contracts := []*Contract{}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("failed to close results: %+v", err)
		}
	}()

	for r.Next() {
		contract := &Contract{}
		if err := r.StructScan(contract); err != nil {
			return nil, err
		}
		contracts = append(contracts, contract)
	}

	return getContractItems(ctx, contracts)
}

// getContractItems fills in the Items of each contract passed
func getContractItems(
	ctx context.Context,
	contracts []*Contract,
) ([]*Contract, error) {
	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)

	for _, contract := range contracts {
		r, err := statements[cx.StmtContractItems].Queryx(
			map[string]interface{}{"contract_id": contract.ID},
		)
		if err != nil {
			return nil, err
		}

		defer func() {
			if err := r.Close(); err != nil {
				log.Printf("failed to close results: %+v", err)
			}
		}()

		contract.Items = []Item{}

		for r.Next() {
			item := Item{}
			if err := r.StructScan(&item); err != nil {
				return nil, err
			}
			contract.Items = append(contract.Items, item)
		}

		log.Printf("contract is: %+v", contract)
	}

	return contracts, nil
}

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
