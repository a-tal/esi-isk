package db

import (
	"context"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
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
	Issued time.Time `db:"issued" json:"issued"`

	// Expires timestamp
	Expires time.Time `db:"expires" json:"expires"`

	// Accepted boolean
	Accepted bool `db:"accepted" json:"accepted"`

	// TODO
	// LocationName is the resolved location name
	// LocationName string

	// System is the solar system ID
	System int32 `db:"system" json:"system"`

	// Items is an array of items in the contract
	Items []*Item `json:"items"`
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
	rows, err := queryNamedResult(ctx, key, map[string]interface{}{
		"character_id": charID,
	})
	if err != nil {
		return nil, err
	}

	res, err := scan(rows, func() interface{} { return &Contract{} })
	if err != nil {
		return nil, err
	}
	contracts := []*Contract{}
	for _, i := range res {
		contracts = append(contracts, i.(*Contract))
	}

	return getContractItems(ctx, contracts)
}

// getContractItems fills in the Items of each contract passed
func getContractItems(
	ctx context.Context,
	contracts []*Contract,
) ([]*Contract, error) {
	for _, contract := range contracts {
		rows, err := queryNamedResult(
			ctx,
			cx.StmtContractItems,
			map[string]interface{}{"contract_id": contract.ID},
		)

		if err != nil {
			return nil, err
		}

		res, err := scan(rows, func() interface{} { return &Item{} })
		if err != nil {
			return nil, err
		}

		contract.Items = []*Item{}
		for _, i := range res {
			contract.Items = append(contract.Items, i.(*Item))
		}
	}

	return contracts, nil
}
