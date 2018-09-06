package db

import (
	"context"
	"fmt"

	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/jmoiron/sqlx"
)

// Name represents an ID -> name mapping
type Name struct {
	ID   int32  `db:"id"`
	Name string `db:"name"`
}

// SaveNames stores the map of ids:names in the db
func SaveNames(ctx context.Context, affiliations []*Affiliation) error {
	names := map[int32]string{}
	for _, aff := range affiliations {
		for _, name := range []*Name{aff.Character, aff.Corporation, aff.Alliance} {
			if name == nil {
				continue
			}
			if _, found := names[name.ID]; found {
				continue
			}
			names[name.ID] = name.Name
		}
	}
	return saveNames(ctx, names)
}

func saveNames(ctx context.Context, names map[int32]string) error {
	known := map[int32]string{}
	for id := range names {
		knownNames, err := GetNames(ctx, id)
		if err == nil {
			for knownID, knownName := range knownNames {
				known[knownID] = knownName
			}
		}
	}

	for id, name := range names {
		prev, found := known[id]
		if found {
			if prev != name {
				if err := updateName(ctx, id, name); err != nil {
					return err
				}
			}
		} else {
			if err := newName(ctx, id, name); err != nil {
				return err
			}
		}
	}

	return nil
}

func newName(ctx context.Context, id int32, name string) error {
	return executeNamed(
		ctx,
		cx.StmtNewName,
		map[string]interface{}{
			"id":   id,
			"name": name,
		},
	)
}

func updateName(ctx context.Context, id int32, name string) error {
	return executeNamed(
		ctx,
		cx.StmtUpdateName,
		map[string]interface{}{
			"id":   id,
			"name": name,
		},
	)
}

// GetNames returns the names for the IDs from the db
func GetNames(ctx context.Context, ids ...int32) (map[int32]string, error) {
	names := map[int32]string{}

	for _, id := range ids {
		name, err := getName(ctx, id)
		if err != nil {
			return nil, err
		}
		names[id] = name
	}

	return names, nil
}

func getName(ctx context.Context, id int32) (string, error) {
	res, err := queryNamedResult(
		ctx,
		cx.StmtGetName,
		map[string]interface{}{"id": id},
	)
	if err != nil {
		return "", err
	}

	names, err := scanNames(res)
	if err != nil {
		return "", err
	}

	for _, name := range names {
		return name.Name, nil
	}

	return "", fmt.Errorf("name for %d was not found", id)
}

func scanNames(rows *sqlx.Rows) ([]*Name, error) {
	names := []*Name{}

	for rows.Next() {
		name := &Name{}
		if err := rows.StructScan(name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}
