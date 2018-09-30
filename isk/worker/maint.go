package worker

import (
	"context"
	"log"

	"github.com/a-tal/esi-isk/isk/db"
)

func pruneContracts(ctx context.Context) {
	contracts, err := db.GetStaleContracts(ctx)
	if err != nil {
		log.Printf("failed to get stale contracts: %+v", err)
		return
	}

	for _, contract := range contracts {
		if err := db.PruneContract(ctx, contract); err != nil {
			log.Printf("failed to prune stale contract: %+v", err)
		}
	}

	if len(contracts) > 0 {
		aff := getContractNames(ctx, contracts)
		if err := db.SaveCharacterContracts(ctx, contracts, aff, false); err != nil {
			log.Printf("failed to save contracts after pruning: %+v", err)
		}
		log.Printf("pruned %d contracts", len(contracts))
	}
}

func pruneDonations(ctx context.Context) {
	donations, err := db.GetStaleDonations(ctx)
	if err != nil {
		log.Printf("failed to get stale donations: %+v", err)
		return
	}

	for _, donation := range donations {
		if err := db.PruneDonation(ctx, donation); err != nil {
			log.Printf("failed to prune stale donation: %+v", err)
		}
	}

	if len(donations) > 0 {
		aff := getNames(ctx, donations)
		if err := db.SaveCharacterDonations(ctx, donations, aff, false); err != nil {
			log.Printf("failed to save donations after pruning: %+v", err)
		}
		log.Printf("pruned %d donations", len(donations))
	}
}
