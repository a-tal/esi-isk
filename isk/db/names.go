package db

import (
	"context"

	"github.com/a-tal/esi-isk/isk/cx"
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
	name := &Name{}
	values := map[string]interface{}{"id": id}
	if err := getNamedResult(ctx, cx.StmtGetName, name, values); err != nil {
		return "", err
	}
	return name.Name, nil
}
