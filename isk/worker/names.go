package worker

import (
	"context"
	"log"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/db"
)

// getNames for all characters involved in the donations
func getNames(
	ctx context.Context,
	donations db.Donations,
) []*db.Affiliation {
	affiliations := []*db.Affiliation{}
	for _, donation := range donations {
		for _, charID := range []int32{donation.Donator, donation.Recipient} {
			isKnown := false
			for _, known := range affiliations {
				if known.Character.ID == charID || known.Corporation.ID == charID {
					isKnown = true
				}
			}
			if isKnown {
				continue
			}

			resolved, err := resolveNames(ctx, charID)
			if err != nil {
				log.Printf("failed to resolve names: %+v", err)
			} else {
				affiliations = append(affiliations, resolved)
			}
		}
	}
	return affiliations
}

// getContractNames for all characters involved in the contracts
func getContractNames(
	ctx context.Context,
	contracts db.Contracts,
) []*db.Affiliation {
	affiliations := []*db.Affiliation{}
	for _, contract := range contracts {
		for _, charID := range []int32{contract.Donator, contract.Receiver} {
			isKnown := false
			for _, known := range affiliations {
				if known.Character.ID == charID || known.Corporation.ID == charID {
					isKnown = true
				}
			}
			if isKnown {
				continue
			}

			resolved, err := resolveNames(ctx, charID)
			if err != nil {
				log.Printf("failed to resolve names: %+v", err)
			} else {
				affiliations = append(affiliations, resolved)
			}
		}
	}
	return affiliations
}

// resolveNames gets the name of the charID,
// which might be a corp or allianceID. it will
// also resolve upwards, so corp+alliance in case
// of charID, and allianceID in case of corp
func resolveNames(ctx context.Context, charID int32) (*db.Affiliation, error) {

	aff := &db.Affiliation{}

	ret, err := ResolveName(ctx, charID)
	if err != nil {
		return nil, err
	}

	for _, res := range ret {

		if res.Category == "corporation" {
			aff.Corporation = &db.Name{ID: res.Id, Name: res.Name}

			allianceID, allianceName := ResolveCorporation(ctx, res.Id)
			if allianceID > 0 {
				aff.Alliance = &db.Name{ID: allianceID, Name: allianceName}
			}

		} else if res.Category == "character" {
			aff.Character = &db.Name{ID: res.Id, Name: res.Name}

			corpID, corpName := ResolveCharacter(ctx, res.Id)
			if corpID > 0 { // 0 == error in lookup
				aff.Corporation = &db.Name{ID: corpID, Name: corpName}
				allianceID, allianceName := ResolveCorporation(ctx, corpID)
				if allianceID > 0 { // 0 == error or not in alliance
					aff.Alliance = &db.Name{ID: allianceID, Name: allianceName}
				}
			}

		} else {
			// hopefully this doesn't happen
			log.Printf(
				"character received donation from %d who is a %s",
				charID,
				res.Category,
			)
		}
	}

	return aff, nil
}

// ResolveCorporation returns the ID and name of the corporation's alliance
func ResolveCorporation(ctx context.Context, corpID int32) (int32, string) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)

	ret, _, err := client.ESI.CorporationApi.GetCorporationsCorporationId(
		ctx,
		corpID,
		nil,
	)
	if err != nil {
		log.Printf("failed to get corporation by ID: %+v", err)
		return 0, ""
	}

	if ret.AllianceId > 0 {
		allianceNameRes, err := ResolveName(ctx, ret.AllianceId)
		if err != nil {
			log.Printf("failed to resolve the alliance of corp: %d", corpID)
			return 0, ""
		}
		for _, res := range allianceNameRes {
			if res.Category == "alliance" && res.Id == ret.AllianceId {
				return ret.AllianceId, res.Name
			}
		}
	}

	return 0, ""
}

// ResolveCharacter returns the ID and name of the character's corporation
func ResolveCharacter(ctx context.Context, charID int32) (int32, string) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)

	ret, _, err := client.ESI.CharacterApi.GetCharactersCharacterId(
		ctx,
		charID,
		nil,
	)

	if err != nil {
		log.Printf("failed to lookup character %d by ID: %+v", charID, err)
		return 0, ""
	}

	corpRes, err := ResolveName(ctx, ret.CorporationId)
	if err != nil {
		log.Printf("failed to resolve name of corp %d: %+v", ret.CorporationId, err)
		return 0, ""
	}

	for _, res := range corpRes {
		if res.Category == "corporation" {
			return res.Id, res.Name
		}
	}

	return 0, ""
}

// ResolveName returns the post universe names return
func ResolveName(ctx context.Context, charID ...int32) (
	[]esi.PostUniverseNames200Ok,
	error,
) {
	client := ctx.Value(cx.Client).(*goesi.APIClient)
	ret, _, err := client.ESI.UniverseApi.PostUniverseNames(ctx, charID, nil)
	return ret, err
}
