package db

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"

	"github.com/a-tal/esi-isk/isk/cx"
)

const (
	// DefaultDonationRow is used if the user has not set a donation row pattern
	DefaultDonationRow = "%CHARACTER% just donated %AMOUNT% ISK!"

	// DefaultContractRow is used if the user has not set a contract row pattern
	DefaultContractRow = "%CHARACTER% just contracted %ITEMS% items worth" +
		" %AMOUNT% ISK!"
)

var (
	// RePreferences ensures the row pattern has at least some content
	RePreferences = regexp.MustCompile(`[^( \t)]+`)
)

// Preferences exports Prefs for donations, contracts, or both
// NB: the JSON form of this is only used for the combined view
type Preferences struct {
	Donations *Prefs `json:"donations"`
	Contracts *Prefs `json:"contracts"`
}

// Prefs exports preferences for either donations or contracts
type Prefs struct {
	Header     string  `json:"header,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	Pattern    string  `json:"pattern"`
	Passphrase string  `json:"passphrase,omitempty"`
	Rows       int     `json:"rows"`
	MaxAge     int     `json:"max_age,omitempty"` // seconds
	Minimum    float64 `json:"minimum"`
}

type dbPreferences struct {
	CharacterID             int32          `db:"character_id"`
	DonationRows            int32          `db:"donation_rows"`
	ContractRows            int32          `db:"contract_rows"`
	CombinedRows            int32          `db:"combined_rows"`
	DonationMaxAge          int32          `db:"donation_max_age"`
	ContractMaxAge          int32          `db:"contract_max_age"`
	CombinedMaxAge          int32          `db:"combined_max_age"`
	DonationMinimum         float64        `db:"donation_min"`
	ContractMinimum         float64        `db:"contract_min"`
	CombinedMinimumDonation float64        `db:"combined_min_donation"`
	CombinedMinimumContract float64        `db:"combined_min_contract"`
	DonationHeader          sql.NullString `db:"donation_header"`
	DonationFooter          sql.NullString `db:"donation_footer"`
	DonationPattern         sql.NullString `db:"donation_pattern"`
	ContractHeader          sql.NullString `db:"contract_header"`
	ContractFooter          sql.NullString `db:"contract_footer"`
	ContractPattern         sql.NullString `db:"contract_pattern"`
	CombinedHeader          sql.NullString `db:"combined_header"`
	CombinedFooter          sql.NullString `db:"combined_footer"`
	CombinedDonationPattern sql.NullString `db:"combined_donation_pattern"`
	CombinedContractPattern sql.NullString `db:"combined_contract_pattern"`
	DonationPassphrase      sql.NullString `db:"donation_passphrase"`
	ContractPassphrase      sql.NullString `db:"contract_passphrase"`
	CombinedPassphrase      sql.NullString `db:"combined_passphrase"`
}

// UserError can bubble up http errors to the api package
type UserError struct {
	Msg  []byte
	Code int
}

func (e UserError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Msg)
}

func getPattern(pattern sql.NullString, fallback string) string {
	if pattern.Valid && RePreferences.MatchString(pattern.String) {
		return pattern.String
	}
	return fallback
}

func getRows(ctx context.Context, rows int32) int {
	intRows := int(rows)
	opts := ctx.Value(cx.Opts).(*cx.Options)
	if intRows > opts.MaxPrefRows {
		return opts.MaxPrefRows
	} else if intRows < 1 {
		return 1
	}
	return intRows
}

func (p *dbPreferences) toPreferences(ctx context.Context, t string) (
	*Preferences,
	error,
) {
	switch t {
	case "d", "":
		return &Preferences{
			Donations: &Prefs{
				Header:     p.DonationHeader.String,
				Footer:     p.DonationFooter.String,
				Pattern:    getPattern(p.DonationPattern, DefaultDonationRow),
				Rows:       getRows(ctx, p.DonationRows),
				Minimum:    p.DonationMinimum,
				Passphrase: p.DonationPassphrase.String,
				MaxAge:     int(p.DonationMaxAge),
			},
		}, nil

	case "c":
		return &Preferences{
			Contracts: &Prefs{
				Header:     p.ContractHeader.String,
				Footer:     p.ContractFooter.String,
				Pattern:    getPattern(p.ContractPattern, DefaultContractRow),
				Rows:       getRows(ctx, p.ContractRows),
				Minimum:    p.ContractMinimum,
				Passphrase: p.ContractPassphrase.String,
				MaxAge:     int(p.ContractMaxAge),
			},
		}, nil

	case "a":
		return &Preferences{
			Donations: &Prefs{
				Header:     p.CombinedHeader.String,
				Footer:     p.CombinedFooter.String,
				Pattern:    getPattern(p.CombinedDonationPattern, DefaultDonationRow),
				Rows:       getRows(ctx, p.CombinedRows),
				Minimum:    p.CombinedMinimumDonation,
				Passphrase: p.CombinedPassphrase.String,
				MaxAge:     int(p.CombinedMaxAge),
			},
			Contracts: &Prefs{
				Header:     p.CombinedHeader.String,
				Footer:     p.CombinedFooter.String,
				Pattern:    getPattern(p.CombinedContractPattern, DefaultContractRow),
				Rows:       getRows(ctx, p.CombinedRows),
				Minimum:    p.CombinedMinimumContract,
				Passphrase: p.CombinedPassphrase.String,
				MaxAge:     int(p.CombinedMaxAge),
			},
		}, nil

	default:
		return nil, UserError{
			Msg:  []byte("Unknown preference type"),
			Code: 400,
		}

	}
}

// GetPreferences returns the Preferences for the logged in user
func GetPreferences(ctx context.Context, t string, charID int32) (
	*Preferences,
	error,
) {
	dbp, err := dbPrefs(ctx, charID)
	if err != nil {
		return nil, err
	}

	p, err := dbp.toPreferences(ctx, t)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// dbPrefs pulls the preferences from the database
func dbPrefs(ctx context.Context, charID int32) (*dbPreferences, error) {
	rows, err := queryNamedResult(
		ctx,
		cx.StmtGetPreferences,
		map[string]interface{}{"character_id": charID},
	)

	if err != nil {
		return nil, err
	}

	res, err := scan(rows, func() interface{} { return &dbPreferences{} })
	if err != nil {
		return nil, err
	}

	for _, i := range res {
		return i.(*dbPreferences), nil
	}

	return nil, UserError{
		Msg:  []byte("Unknown character ID"),
		Code: 500,
	}
}

// SetPreferences sets the Preferences for the logged in user
func SetPreferences(ctx context.Context, charID int32, p *Preferences) error {
	if p.Contracts != nil && p.Donations != nil {
		return setPreferences(ctx, charID, p)
	} else if p.Contracts != nil {
		return setPrefs(ctx, charID, p.Contracts, cx.StmtSetContractPreferences)
	} else {
		return setPrefs(ctx, charID, p.Donations, cx.StmtSetDonationPreferences)
	}
}

// setPreferences stores combined preferences
func setPreferences(ctx context.Context, charID int32, p *Preferences) error {
	return executeNamed(
		ctx,
		cx.StmtSetCombinedPreferences,
		map[string]interface{}{
			"character_id":     charID,
			"header":           p.Donations.Header,
			"footer":           p.Donations.Footer,
			"rows":             p.Donations.Rows,
			"max_age":          p.Donations.MaxAge,
			"donation_pattern": p.Donations.Pattern,
			"contract_pattern": p.Contracts.Pattern,
			"donation_minimum": p.Donations.Minimum,
			"contract_minimum": p.Contracts.Minimum,
			"passphrase":       p.Donations.Passphrase,
		},
	)
}

func setPrefs(ctx context.Context, charID int32, p *Prefs, key cx.Key) error {
	return executeNamed(
		ctx,
		key,
		map[string]interface{}{
			"character_id": charID,
			"header":       p.Header,
			"footer":       p.Footer,
			"pattern":      p.Pattern,
			"rows":         p.Rows,
			"minimum":      p.Minimum,
			"max_age":      p.MaxAge,
			"passphrase":   p.Passphrase,
		},
	)
}

// Sanity ensures our attribute lengths are acceptable
func (p *Prefs) Sanity(ctx context.Context) error {
	p.Rows = getRows(ctx, int32(p.Rows))

	opts := ctx.Value(cx.Opts).(*cx.Options)
	if stringLen(p.Header) > opts.MaxPrefLen ||
		stringLen(p.Footer) > opts.MaxPrefLen ||
		stringLen(p.Pattern) > opts.MaxPatternLen {
		return UserError{
			Msg:  []byte("Preference string too long"),
			Code: 400,
		}
	}

	return nil
}

func stringLen(s string) int32 {
	l := len(s)
	if l > math.MaxInt32 {
		return int32(math.MaxInt32)
	}
	return int32(l)
}
