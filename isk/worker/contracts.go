package worker

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"

	"github.com/a-tal/esi-isk/isk/api"
	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
)

// marketPrices stores market prices from ESI in memory
type marketPrices struct {
	lock    *sync.Mutex
	prices  map[int32]float64
	expires time.Time
}

func newPrices(ctx context.Context) (*marketPrices, error) {
	prices, expires, err := getPrices(ctx)
	if err != nil {
		return nil, err
	}

	m := &marketPrices{
		lock:    &sync.Mutex{},
		prices:  prices,
		expires: expires,
	}
	go m.updater(ctx)
	return m, nil
}

func (m *marketPrices) value(items map[int32]int32) float64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	sum := float64(0)
	for typeID, quantity := range items {
		sum += (m.prices[typeID] * float64(quantity))
	}
	return sum
}

func (m *marketPrices) updater(ctx context.Context) {
	minDt := time.Duration(60 * time.Second)

	for {
		dt := m.expires.Sub(time.Now().UTC())
		if dt < minDt {
			dt = minDt
		}

		log.Printf("next market update in: %+v", dt)
		time.Sleep(dt)
		log.Println("updating market prices")

		prices, expires, err := getPrices(ctx)
		if err != nil {
			log.Printf("failed to update market prices: %+v", err)
		} else {
			m.lock.Lock()
			m.prices = prices
			m.expires = expires
			m.lock.Unlock()
		}
	}
}

func getPrices(ctx context.Context) (
	prices map[int32]float64,
	expires time.Time,
	err error,
) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	esiPrices, r, err := client.ESI.MarketApi.GetMarketsPrices(ctx, nil)
	if err != nil {
		return
	}

	expires, err = getExpires(r)
	if err != nil {
		return
	}

	prices = map[int32]float64{}
	for _, i := range esiPrices {
		prices[i.TypeId] = i.AdjustedPrice
	}

	return prices, expires, nil
}

// pull the next update time from the response headers
func getExpires(r *http.Response) (expires time.Time, err error) {
	expires, err = time.Parse(api.RFC1123, r.Header.Get("Expires"))
	if err != nil {
		return
	}

	return expires.Add(1 * time.Second), nil
}

func characterContracts(ctx context.Context, user *db.User) ([]int32, error) {
	charIDs := []int32{}

	contracts, err := getContracts(ctx, user)
	if err != nil {
		return charIDs, err
	}

	sort.Sort(contracts)

	prevID := int32(user.LastContractID.Int64)

	setLastContractID(contracts, user)

	outstanding, err := db.GetOutstandingContracts(ctx, user.CharacterID)
	if err != nil {
		return charIDs, err
	}

	new, updated := parseForZeroISK(contracts, user, prevID, outstanding)
	donations, updates := asDbContracts(ctx, new, updated)

	if len(donations) > 0 {
		charIDs = append(charIDs, user.CharacterID)
	}

	for _, donation := range donations {
		charIDs = append(charIDs, donation.Donator)
	}

	return charIDs, saveContractRun(
		ctx,
		donations,
		updates,
		getContractNames(ctx, donations),
	)
}

func saveContractRun(
	ctx context.Context,
	contracts []*db.Contract,
	updates []*db.Contract,
	affiliations []*db.Affiliation,
) error {
	for _, contract := range contracts {
		if err := db.SaveContract(ctx, contract); err != nil {
			return err
		}
	}

	if err := db.UpdateContracts(ctx, updates, affiliations); err != nil {
		return err
	}

	if err := db.SaveNames(ctx, affiliations); err != nil {
		return err
	}

	return db.SaveCharacterContracts(ctx, contracts, affiliations, true)
}

func getContractValue(ctx context.Context, items []*db.Item) float64 {
	itemQuantities := map[int32]int32{}
	for _, item := range items {
		itemQuantities[item.TypeID] += item.Quantity
	}

	m := ctx.Value(cx.Prices).(*marketPrices)
	return m.value(itemQuantities)
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
		if hasLastID && contract.ContractId == lastID {
			return true
		}
	}
	return false
}

func getLastContractID(user *db.User) (bool, int32) {
	return user.LastContractID.Valid, int32(user.LastContractID.Int64)
}

// parseForZeroISK finds contracts that are zero ISK item exchanges
func parseForZeroISK(
	contracts []esi.GetCharactersCharacterIdContracts200Ok,
	user *db.User,
	prevID int32,
	outstanding []int32,
) (
	new []esi.GetCharactersCharacterIdContracts200Ok,
	updated []esi.GetCharactersCharacterIdContracts200Ok,
) {
	new = []esi.GetCharactersCharacterIdContracts200Ok{}
	updated = []esi.GetCharactersCharacterIdContracts200Ok{}
	newContracts := true
	for _, contract := range contracts {
		if contract.ContractId == prevID {
			newContracts = false
		}
		if contract.Type_ == "item_exchange" && contract.Price == 0 {
			if newContracts {
				new = append(new, contract)
			} else {
				for i := 0; i < len(outstanding); i++ {
					if outstanding[i] == contract.ContractId {
						if contract.Status != "outstanding" {
							log.Printf("outstanding contract %d has updated", contract.ContractId)
							updated = append(updated, contract)
						} else {
							log.Printf("contract %d is still outstanding", contract.ContractId)
						}
					}
				}
			}
		}
	}
	return new, updated
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

func toDbContract(c esi.GetCharactersCharacterIdContracts200Ok) *db.Contract {
	return &db.Contract{
		ID:       c.ContractId,
		Donator:  c.IssuerId,
		Receiver: c.AssigneeId,
		Location: c.StartLocationId,
		Issued:   c.DateIssued,
		Expires:  c.DateExpired,
		Accepted: c.Status == "finished",
		Note:     c.Title,
	}
}

// asDbContracts fills in Items and Value and converts into *db.Contract
func asDbContracts(
	ctx context.Context,
	contracts zeroISKContracts,
	updates zeroISKContracts,
) ([]*db.Contract, []*db.Contract) {
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

		c := toDbContract(contract)
		c.Value = getContractValue(ctx, items)
		c.Items = items

		zeroISK = append(zeroISK, c)
	}

	updateContracts := []*db.Contract{}
	for _, update := range updates {
		updateContracts = append(updateContracts, toDbContract(update))
	}

	return zeroISK, updateContracts
}
