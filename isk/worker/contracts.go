package worker

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
)

func pullCharacterContracts(ctx context.Context, user *db.User) error {
	// TODO need to check for contract's accepted status changing

	contracts, err := getContracts(ctx, user)
	if err != nil {
		return err
	}

	sort.Sort(contracts)

	setLastContractID(contracts, user)

	donations := asDbContracts(ctx, parseForZeroISK(contracts, user))

	// XXX should check all unaccepted contracts for this character
	//     and update their status if changed. needs new query

	return saveContractRun(ctx, donations, getContractNames(ctx, user, donations))
}

func saveContractRun(
	ctx context.Context,
	contracts []*db.Contract,
	affiliations []*db.Affiliation,
) error {
	for _, contract := range contracts {
		if err := db.SaveContract(ctx, contract); err != nil {
			return err
		}
	}

	if err := db.SaveNames(ctx, affiliations); err != nil {
		return err
	}

	return db.SaveCharacterContracts(ctx, contracts, affiliations)
}

func getContractValue(items []*db.Item) float64 {
	// TODO: pull value from evepraisal once they support item id posting
	// https://github.com/evepraisal/go-evepraisal/issues/81
	log.Println("todo: lookup price of %+v", items)
	return 666.66
}

func getContractItems(
	ctx context.Context,
	contract esi.GetCharactersCharacterIdContracts200Ok,
) ([]*db.Item, error) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	api := client.ESI.ContractsApi

	items, _, err := api.GetCharactersCharacterIdContractsContractIdItems(
		ctx,
		contract.AssigneeId,
		contract.ContractId,
		nil,
	)
	if err != nil {
		return nil, err
	}

	dbItems := []*db.Item{}
	for _, item := range items {
		dbItems = append(dbItems, &db.Item{
			ID:         item.RecordId,
			ContractID: contract.ContractId,
			TypeID:     item.TypeId,
			Quantity:   item.Quantity,
			// ItemID: item.ItemId,
		})
	}

	return dbItems, nil
}

func getContracts(ctx context.Context, user *db.User) (
	zeroISKContracts,
	error,
) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	entries, r, err := client.ESI.ContractsApi.GetCharactersCharacterIdContracts(
		ctx,
		user.CharacterID,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if !knownContract(entries, user) {
		additional, err := expandContracts(ctx, user, r)
		if err != nil {
			return nil, err
		}
		entries = append(entries, additional...)
	}

	return entries, nil
}

// return true if we've seen one or more of these contracts
func knownContract(
	contracts []esi.GetCharactersCharacterIdContracts200Ok,
	user *db.User,
) bool {
	hasLastID, lastID := getLastContractID(user)
	for _, contract := range contracts {
		if hasLastID && int64(contract.ContractId) == lastID {
			return true
		}
	}
	return false
}

func getLastContractID(user *db.User) (bool, int64) {
	return user.LastContractID.Valid, user.LastContractID.Int64
}

// parseForZeroISK finds contracts that are zero ISK item exchanges
func parseForZeroISK(
	contracts []esi.GetCharactersCharacterIdContracts200Ok,
	user *db.User,
) []esi.GetCharactersCharacterIdContracts200Ok {
	zeroISK := []esi.GetCharactersCharacterIdContracts200Ok{}
	for _, contract := range contracts {
		if contract.Type_ == "item_exchange" && contract.Price == 0 {
			zeroISK = append(zeroISK, contract)
		}
	}
	return zeroISK
}

func setLastContractID(contracts zeroISKContracts, user *db.User) {
	if len(contracts) < 1 {
		return
	}
	user.LastContractID = sql.NullInt64{
		Int64: int64(contracts[0].ContractId),
		Valid: true,
	}
}

type zeroISKContracts []esi.GetCharactersCharacterIdContracts200Ok

func (z zeroISKContracts) Len() int      { return len(z) }
func (z zeroISKContracts) Swap(i, j int) { z[i], z[j] = z[j], z[i] }
func (z zeroISKContracts) Less(i, j int) bool {
	return z[i].DateIssued.Before(z[j].DateIssued)
}

func expandContracts(
	ctx context.Context,
	user *db.User,
	res *http.Response,
) (zeroISKContracts, error) {
	additional := zeroISKContracts{}
	xPagesRaw := res.Header.Get("X-Pages")
	if xPagesRaw == "" {
		return additional, nil
	}

	xPages64, err := strconv.ParseInt(xPagesRaw, 10, 32)
	if err != nil {
		return additional, err
	}
	xPages := int(xPages64)

	more := make(chan zeroISKContracts)
	errs := make(chan error)

	defer close(more)
	defer close(errs)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	started := 0
	for i := 2; i <= xPages; i++ {
		started++
		go func(i int) {
			entries, additionalErr := additionalContractPage(ctx, user, int32(i))
			if additionalErr != nil {
				errs <- additionalErr
				return
			}
			more <- entries
		}(i)
	}

	for i := 0; i < started; i++ {
		select {
		case entries := <-more:
			additional = append(additional, entries...)
		case err = <-errs:
			cancel()
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return additional, nil
}

func additionalContractPage(
	ctx context.Context,
	user *db.User,
	page int32,
) (zeroISKContracts, error) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	entries, _, err := client.ESI.ContractsApi.GetCharactersCharacterIdContracts(
		ctx,
		user.CharacterID,
		&esi.GetCharactersCharacterIdContractsOpts{
			Page: optional.NewInt32(page),
		},
	)
	return entries, err
}

// asDbContracts fills in Items and Value and converts into *db.Contract
func asDbContracts(
	ctx context.Context,
	contracts zeroISKContracts,
) []*db.Contract {
	zeroISK := []*db.Contract{}

	for _, contract := range contracts {
		items, err := getContractItems(ctx, contract)
		if err != nil {
			log.Printf(
				"failed to lookup contract items for %d: %+v",
				contract.ContractId,
				err,
			)
			continue
		}

		zeroISK = append(zeroISK, &db.Contract{
			ID:       contract.ContractId,
			Donator:  contract.IssuerId,
			Receiver: contract.AssigneeId,
			Location: contract.StartLocationId,
			Issued:   contract.DateIssued,
			Expires:  contract.DateExpired,
			Accepted: contract.Status == "finished",
			Value:    getContractValue(items),
			Note:     contract.Title,
			Items:    items,
		})
	}

	return zeroISK
}
