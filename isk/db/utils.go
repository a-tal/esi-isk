package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // adds the "postgres" driver to sql

	"github.com/a-tal/esi-isk/isk/cx"
)

// Connect returns a new connection to the postgres db
func Connect(ctx context.Context) *sqlx.DB {
	opts := ctx.Value(cx.Opts).(*cx.Options)
	db, err := sqlx.Open("postgres", fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		opts.DB.User,
		opts.DB.Password,
		opts.DB.Host,
		opts.DB.Name,
		opts.DB.Mode,
	))
	if err != nil {
		log.Fatal(err)
	}

	// open doesn't actually confirm connectivity... do that now
	if pingErr := db.Ping(); pingErr != nil {
		log.Fatal(pingErr)
	}

	// if we're not in test, remove any test data?
	// XXX XXX XXX WHYYYYYYY
	if opts.Production {
		// XXX TODO
		log.Println("TODO: delete test data here")
	}

	log.Println("db connection ok")
	return db
}

func queryNamedResult(
	ctx context.Context,
	stmt cx.Key,
	values map[string]interface{},
) (*sqlx.Rows, error) {
	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)
	return statements[stmt].Queryx(values)
}

func executeNamed(
	ctx context.Context,
	stmt cx.Key,
	values map[string]interface{},
) error {
	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)
	_, err := statements[stmt].Exec(values)
	return err
}
