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

func getNamedResult(
	ctx context.Context,
	stmt cx.Key,
	dest interface{},
	values map[string]interface{},
) error {
	statements := ctx.Value(cx.Statements).(map[cx.Key]*sqlx.NamedStmt)
	return statements[stmt].Get(dest, values)
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

func inInt32(i int32, l []int32) bool {
	for _, j := range l {
		if i == j {
			return true
		}
	}
	return false
}

func scan(rows *sqlx.Rows, newItem func() interface{}) ([]interface{}, error) {
	items := []interface{}{}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %+v", err)
		}
	}()

	for rows.Next() {
		item := newItem()
		if err := rows.StructScan(item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
