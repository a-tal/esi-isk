package db

import (
	"context"
	"sort"
	"time"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/lib/pq"
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
	// TODO: resolve locationID to a name (or system name)

	// Issued timestamp
	Issued time.Time `db:"issued" json:"issued"`

	// Expires timestamp
	Expires time.Time `db:"expires" json:"expires"`

	// Accepted boolean
	Accepted bool `db:"accepted" json:"accepted"`

	// Value is an estimated value of the contract items
	Value float64 `db:"value" json:"value"`

	// Note is the title of the contract
	Note string `db:"note" json:"note"`

	// Items is an array of items in the contract
	Items []*Item `json:"items"`
}

// Contracts are time sorted
type Contracts []*Contract

func (c Contracts) Len() int      { return len(c) }
func (c Contracts) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c Contracts) Less(i, j int) bool {
	return c[i].Issued.After(c[j].Issued)
}

// Item are sourced from the contractItems table by id
type Item struct {
	// ID is a record ID
	ID int64 `db:"id" json:"-"`

	// ContractID links this item to a contract
	ContractID int32 `db:"contract_id" json:"-"`

	// TypeID of the item in the contract
	TypeID int32 `db:"type_id" json:"type_id"`

	// Quantity of items given
	Quantity int32 `db:"quantity" json:"quantity"`

	// ItemID of the item in the contract (if possible to determine)
	ItemID int64 `db:"item_id" json:"item_id,omitempty"`
}

func getCharContracts(ctx context.Context, charID int32) (Contracts, error) {
	return getContracts(ctx, charID, cx.StmtCharContracts)
}

func getCharContracted(ctx context.Context, charID int32) (Contracts, error) {
	return getContracts(ctx, charID, cx.StmtCharContracted)
}

func getContracts(ctx context.Context, charID int32, key cx.Key) (
	Contracts,
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
	contracts := Contracts{}
	for _, i := range res {
		c := i.(*Contract)
		c.Value = round2(c.Value)
		contracts = append(contracts, c)
	}

	sort.Sort(contracts)

	itemErr := GetContractItems(ctx, contracts)
	return contracts, itemErr
}

// GetStaleContracts returns contracts issued more than 30 days ago
func GetStaleContracts(ctx context.Context) (Contracts, error) {
	rows, err := queryNamedResult(ctx, cx.StmtGetStaleContracts, nil)
	if err != nil {
		return nil, err
	}

	res, err := scan(rows, func() interface{} { return &Contract{} })
	if err != nil {
		return nil, err
	}
	contracts := Contracts{}
	for _, i := range res {
		contracts = append(contracts, i.(*Contract))
	}

	return contracts, nil
}

// PruneContract removes a contract and deducts from the 30d totals
func PruneContract(ctx context.Context, c *Contract) error {
	if err := executeContract(ctx, cx.StmtRemoveContract, c); err != nil {
		return err
	}
	return executeContract(ctx, cx.StmtRemoveContractItems, c)
}

func executeContract(ctx context.Context, key cx.Key, c *Contract) error {
	return executeNamed(ctx, key, map[string]interface{}{"contract_id": c.ID})
}

// GetOutstandingContracts returns a list of outstanding contract IDs
func GetOutstandingContracts(ctx context.Context, c int32) ([]int32, error) {
	contracts, err := getContracts(ctx, c, cx.StmtGetOutstandingContracts)
	if err != nil {
		return nil, err
	}
	outstandingIDs := []int32{}
	for _, contract := range contracts {
		outstandingIDs = append(outstandingIDs, contract.ID)
	}
	return outstandingIDs, nil
}

// GetContractItems fills in the Items of each Contract passed
func GetContractItems(
	ctx context.Context,
	contracts Contracts,
) error {
	for _, contract := range contracts {
		rows, err := queryNamedResult(
			ctx,
			cx.StmtContractItems,
			map[string]interface{}{"contract_id": contract.ID},
		)

		if err != nil {
			return err
		}

		res, err := scan(rows, func() interface{} { return &Item{} })
		if err != nil {
			return err
		}

		contract.Items = []*Item{}
		for _, i := range res {
			contract.Items = append(contract.Items, i.(*Item))
		}
	}

	return nil
}

// SaveContract saves the contract and associated items in the db
func SaveContract(ctx context.Context, contract *Contract) error {
	err := executeNamed(ctx, cx.StmtAddContract, map[string]interface{}{
		"contract_id": contract.ID,
		"donator":     contract.Donator,
		"receiver":    contract.Receiver,
		"location":    contract.Location,
		"issued":      contract.Issued,
		"expires":     contract.Expires,
		"accepted":    contract.Accepted,
		"value":       contract.Value,
		"note":        contract.Note,
	})
	if err != nil {
		return err
	}
	return saveContractItems(ctx, contract.Items)
}

// UpdateContracts sets the contract as accepted in the db, if it has been
func UpdateContracts(
	ctx context.Context,
	contracts []*Contract,
	aff []*Affiliation,
) error {
	if err := SaveCharacterContracts(ctx, contracts, aff, true); err != nil {
		return err
	}

	for _, contract := range contracts {
		if err := executeNamed(ctx, cx.StmtAcceptContract, map[string]interface{}{
			"contract_id":  contract.ID,
			"character_id": contract.Receiver,
		}); err != nil {
			return err
		}
	}

	return nil
}

func saveContractItems(ctx context.Context, items []*Item) error {
	for _, item := range items {
		err := executeNamed(ctx, cx.StmtAddContractItems, map[string]interface{}{
			"id":          item.ID,
			"contract_id": item.ContractID,
			"type_id":     item.TypeID,
			"item_id":     0, // XXX replace once item IDs are in all contract endpoints
			"quantity":    item.Quantity,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// addToContractTotals adds donation/received totals from contracts
func addToContractTotals(contract *Contract, characters ...[]*CharacterRow) {
	for _, chars := range characters {
		for _, char := range chars {
			if char.ID == contract.Donator {
				char.DonatedISK += contract.Value
				char.Donated++
				char.DonatedISK30 += contract.Value
				char.Donated30++
				if !char.LastDonated.Valid || char.LastDonated.Time.Before(
					contract.Issued) {
					char.LastDonated = pq.NullTime{Time: contract.Issued, Valid: true}
					char.LastDonated.Valid = true
				}
			} else if char.ID == contract.Receiver {
				char.ReceivedISK += contract.Value
				char.Received++
				char.ReceivedISK30 += contract.Value
				char.Received30++
				if !char.LastReceived.Valid || char.LastReceived.Time.Before(
					contract.Issued) {
					char.LastReceived = pq.NullTime{Time: contract.Issued, Valid: true}
					char.LastReceived.Valid = true
				}
			}
		}
	}
}

// removeFromContractTotals removes donation/received totals from contracts
func removeFromContractTotals(contract *Contract, chars ...[]*CharacterRow) {
	for _, characters := range chars {
		for _, char := range characters {
			if char.ID == contract.Donator {
				char.DonatedISK30 -= contract.Value
				char.Donated30--
			} else if char.ID == contract.Receiver {
				char.ReceivedISK30 -= contract.Value
				char.Received30--
			}
		}
	}
}
