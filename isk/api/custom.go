package api

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/a-tal/esi-isk/isk/db"
)

// Custom character view, defined by character preferences
func Custom(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		charID, err := getCharID(r)
		if err != nil || charID < 1 {
			write400(w)
			return
		}

		c, err := db.GetCharDetails(ctx, charID)
		if err != nil {
			log.Printf("failed to get character details: %+v", err)
			write500(w)
			return
		}

		p, err := getPreferences(w, r.WithContext(ctx), charID)
		if err != nil {
			// getPreferences writes any errors
			return
		}

		if pErr := checkPassphrase(r, c, p); pErr != nil {
			write403(w)
			return
		}

		header, rows, footer, err := buildTemplates(c)
		if err != nil {
			write500(w)
			return
		}

		writeCacheHeaders(ctx, w)

		if wErr := writeTemplates(ctx, w, header, rows, footer, c, p); wErr != nil {
			write400(w)
		}
	}
}

func checkPassphrase(
	r *http.Request,
	c *db.CharDetails,
	p *db.Preferences,
) error {
	if !c.Character.GoodStanding {
		return nil
	}

	var passphrase string
	if p.Donations != nil {
		passphrase = p.Donations.Passphrase
	} else {
		passphrase = p.Contracts.Passphrase
	}

	if passphrase != "" && r.URL.Query().Get("p") != passphrase {
		return errors.New("incorrect passphrase")
	}

	return nil
}

func writeTemplates(
	ctx context.Context,
	w http.ResponseWriter,
	header, rows, footer *template.Template,
	c *db.CharDetails,
	p *db.Preferences,
) error {

	if err := writeHeader(w, p, header); err != nil {
		return err
	}

	var rowErr error
	if p.Contracts != nil && p.Donations != nil {
		rowErr = writeRowsMultiple(ctx, w, rows, c, p)
	} else if p.Contracts != nil {
		rowErr = writeRowsSingular(ctx, w, rows, c, p.Contracts, "c")
	} else {
		rowErr = writeRowsSingular(ctx, w, rows, c, p.Donations, "d")
	}

	if rowErr != nil {
		return rowErr
	}

	return writeFooter(w, p, footer)
}

func writeHeader(
	w http.ResponseWriter,
	p *db.Preferences,
	t *template.Template,
) error {
	if p.Donations != nil {
		// donation or combined view
		return t.ExecuteTemplate(w, "T", p.Donations.Header)
	}
	return t.ExecuteTemplate(w, "T", p.Contracts.Header)
}

func writeFooter(
	w http.ResponseWriter,
	p *db.Preferences,
	t *template.Template,
) error {
	if p.Donations != nil {
		// donation or combined view
		return t.ExecuteTemplate(w, "T", p.Donations.Footer)
	}
	return t.ExecuteTemplate(w, "T", p.Contracts.Footer)
}

func writeRowsSingular(
	ctx context.Context,
	w http.ResponseWriter,
	rows *template.Template,
	c *db.CharDetails,
	p *db.Prefs,
	t string,
) error {
	for _, pattern := range getRowPatterns(ctx, c, p, t) {
		if err := rows.ExecuteTemplate(w, "T", pattern.str); err != nil {
			return err
		}
	}
	return nil
}

func writeRowsMultiple(
	ctx context.Context,
	w http.ResponseWriter,
	rows *template.Template,
	c *db.CharDetails,
	p *db.Preferences,
) error {
	rp := rowPatterns{}
	rp = append(rp, getRowPatterns(ctx, c, p.Donations, "d")...)
	rp = append(rp, getRowPatterns(ctx, c, p.Contracts, "c")...)

	sort.Sort(rp)

	for i := 0; i < p.Donations.Rows && i < len(rp); i++ {
		if err := rows.ExecuteTemplate(w, "T", rp[i].str); err != nil {
			return err
		}
	}
	return nil
}

type rowPatterns []*rowPattern

func (r rowPatterns) Len() int           { return len(r) }
func (r rowPatterns) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r rowPatterns) Less(i, j int) bool { return r[i].ts.After(r[j].ts) }

type rowPattern struct {
	str string
	ts  time.Time
}

func getRowPatterns(
	ctx context.Context,
	c *db.CharDetails,
	p *db.Prefs,
	t string,
) rowPatterns {
	patterns := rowPatterns{}
	index := 0
	for i := 0; i < p.Rows; i++ {
		var pattern string
		var err error
		var ts time.Time
		pattern, ts, index, err = getRowPattern(ctx, c, p, t, index)
		if err != nil {
			break
		}
		if p.MaxAge != 0 {
			cutoff := time.Now().UTC().Add(-time.Duration(p.MaxAge) * time.Second)
			if ts.Before(cutoff) {
				break
			}
		}
		patterns = append(patterns, &rowPattern{str: pattern, ts: ts})
		index++
	}
	return patterns
}

func getRowPattern(
	ctx context.Context,
	c *db.CharDetails,
	p *db.Prefs,
	t string,
	i int,
) (string, time.Time, int, error) {
	switch t {

	case "d", "":
		donation, index, err := getValidDonation(c, p, i)
		if err != nil {
			return "", time.Time{}, index, err
		}
		pattern, err := getDonationRow(ctx, c, p, donation)
		return pattern, donation.Timestamp, index, err

	case "c":
		contract, index, err := getValidContract(c, p, i)
		if err != nil {
			return "", time.Time{}, index, err
		}
		pattern, err := getContractRow(ctx, c, p, contract)
		return pattern, contract.Issued, index, err

	default:
		return "", time.Time{}, i, errors.New("unknown preference type")

	}
}

func getValidDonation(c *db.CharDetails, p *db.Prefs, i int) (
	*db.Donation,
	int,
	error,
) {
	for {
		if i >= len(c.Donations) {
			return nil, i, errors.New("index out of bounds")
		}
		if c.Donations[i].Amount >= p.Minimum {
			break
		}
		i++
	}
	return c.Donations[i], i, nil
}

func getValidContract(c *db.CharDetails, p *db.Prefs, i int) (
	*db.Contract,
	int,
	error,
) {
	for {
		if i >= len(c.Contracts) {
			return nil, i, errors.New("index out of bounds")
		}
		if c.Contracts[i].Value >= p.Minimum {
			break
		}
		i++
	}
	return c.Contracts[i], i, nil
}

func stdReplacements(isk float64, t time.Time) map[string]string {
	printer := message.NewPrinter(language.English)
	ampmHour, ampm := asAMPM(t.Hour())
	return map[string]string{
		"%AMOUNT%":       printer.Sprintf("%.2f", isk),
		"%AMOUNTISK%":    printer.Sprintf("%.0f", isk),
		"%AMOUNTRAW%":    fmt.Sprintf("%.2f", isk),
		"%AMOUNTRAWISK%": fmt.Sprintf("%.0f", isk),
		"%DAY%":          fmt.Sprintf("%d", t.Day()),
		"%DAYSUFFIX%":    getNumberSuffix(t.Day()),
		"%MONTH%":        t.Month().String()[:3],
		"%MONTHLONG%":    t.Month().String(),
		"%YEAR%":         fmt.Sprintf("%d", t.Year()),
		"%TIME%": fmt.Sprintf(
			"%02d:%02d",
			t.Hour(),
			t.Minute(),
		),
		"%TIMEFULL%": fmt.Sprintf(
			"%02d:%02d:%02d",
			t.Hour(),
			t.Minute(),
			t.Second(),
		),
		"%TIMEAMPM%": fmt.Sprintf(
			"%02d:%02d %s",
			ampmHour,
			t.Minute(),
			ampm,
		),
		"%TIMEFULLAMPM%": fmt.Sprintf(
			"%02d:%02d:%02d %s",
			ampmHour,
			t.Minute(),
			t.Second(),
			ampm,
		),
		"%ISODATE%": t.Format(time.RFC3339),
	}
}

func getDonationRow(
	ctx context.Context,
	c *db.CharDetails,
	p *db.Prefs,
	d *db.Donation,
) (string, error) {
	donator, err := db.GetName(ctx, d.Donator)
	if err != nil {
		return "", err
	}

	replacements := stdReplacements(d.Amount, d.Timestamp)
	replacements["%NAME%"] = c.Character.Name
	replacements["%CHARACTER%"] = donator
	replacements["%NOTE%"] = d.Note

	pattern := p.Pattern
	for search, replace := range replacements {
		pattern = strings.Replace(pattern, search, replace, -1)
	}

	return pattern, nil
}

func asAMPM(hour int) (int, string) {
	if hour > 12 {
		return hour - 12, "PM"
	}
	return hour, "AM"
}

func getNumberSuffix(n int) string {
	if n != 11 && n%10 == 1 {
		return "st"
	} else if n != 12 && n%10 == 2 {
		return "nd"
	} else if n != 13 && n%10 == 3 {
		return "rd"
	}
	return "th"
}

func getContractRow(
	ctx context.Context,
	c *db.CharDetails,
	p *db.Prefs,
	k *db.Contract,
) (string, error) {
	contractor, err := db.GetName(ctx, k.Donator)
	if err != nil {
		return "", err
	}

	replacements := stdReplacements(k.Value, k.Issued)
	replacements["%NAME%"] = c.Character.Name
	replacements["%CHARACTER%"] = contractor
	replacements["%NOTE%"] = k.Note
	replacements["%ITEMS%"] = fmt.Sprintf("%d", len(k.Items))

	pattern := p.Pattern
	for search, replace := range replacements {
		pattern = strings.Replace(pattern, search, replace, -1)
	}

	return pattern, nil
}

func buildTemplates(c *db.CharDetails) (
	header, rows, footer *template.Template,
	err error,
) {
	header, err = template.New("header").Parse(
		fmt.Sprintf(
			`{{define "T"}}<!doctype html>
<html lang="en">
 <head>
  <meta charset="utf-8">
  <meta http-equiv="refresh" content="300">
  <title>ESI ISK - %s</title>
 </head>
 <body>
  <header>{{.}}</header>
  <main>{{end}}`,
			c.Character.Name,
		),
	)

	if err != nil {
		return
	}

	rows, err = template.New("rows").Parse(`{{define "T"}}
   <article>{{.}}</article>{{end}}`,
	)

	if err != nil {
		return
	}

	footer, err = template.New("footer").Parse(`{{define "T"}}
  </main>
  <footer>{{.}}</footer>
 </body>
</html>{{end}}`,
	)

	return header, rows, footer, err
}
