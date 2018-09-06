package db

import (
	"context"
	"log"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/jmoiron/sqlx"
)

// Contract describes zero ISK donation contracts
type Contract struct {
	// ID is the contract ID
	ID int32 `db:"contract_id" json:"id"`

	// Donator who sent the recipient the contract
	Donator int32 `db:"donator" json:"donator"`

	// Receiver is who received the contract
	Receiver int32 `db:"receiver" json:"receiver"`

	// Location is the station or structure ID
	Location int64 `db:"location" json:"location"`

	// Issued timestamp
	Issued *time.Time `db:"issued" json:"issued"`

	// Expires timestamp
	Expires *time.Time `db:"expires" json:"expires"`

	// Accepted boolean
	Accepted bool `db:"accepted" json:"accepted"`

	// TODO
	// LocationName is the resolved location name
	// LocationName string

	// System is the solar system ID
	System int32 `db:"system" json:"system"`

	// Items is an array of items in the contract
	Items []Item `json:"items"`
}

// Item are sourced from the contractItems table by id
type Item struct {
	// ID is an auto-incrementing ID
	// not sure if serial type is 32 or 64b
	ID int64 `db:"id" json:"id"`

	// ContractID links this item to a contract
	ContractID int32 `db:"contract_id" json:"contract_id"`

	// TypeID of the item in the contract
	TypeID int32 `db:"type_id" json:"type_id"`

	// ItemID of the item in the contract (if possible to determine)
	ItemID int64 `db:"item_id" json:"item_id,omitempty"`

	// Quantity of items given
	Quantity int64 `db:"quantity" json:"quantity"`

	// TODO
	// CostPer item in the contract (estimate)
	// CostPer float64
}

func getCharContracts(ctx context.Context, charID int32) ([]*Contract, error) {
	return getContracts(ctx, charID, cx.StmtCharContracts)
}

func getCharContracted(ctx context.Context, charID int32) ([]*Contract, error) {
	return getContracts(ctx, charID, cx.StmtCharContracted)
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
