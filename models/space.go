package models

import (
	"database/sql"
	"github.com/google/uuid"
)

type Space struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Map   string    `json:"map"`
	ModId uuid.UUID `json:"modId,omitempty"`
}

var cachedSpaces = make([]Space, 0)

func GetCachedSpaceById(id uuid.UUID) *Space {
	for _, v := range cachedSpaces {
		if v.Id == id {
			return &v
		}
	}
	return nil
}

func GetSpaceById(db *sql.DB, spaceId uuid.UUID) (*Space, error) {

	//region cache
	cachedSpace := GetCachedSpaceById(spaceId)
	if cachedSpace != nil {
		return cachedSpace, nil
	}
	//endregion

	space := Space{}

	err := db.QueryRow(
		"SELECT e.id, s.name, s.map, s.mod_id FROM spaces AS s LEFT JOIN entities AS e ON s.id=e.id WHERE s.id = $1",
		spaceId,
	).Scan(&space.Id, &space.Name, &space.Map, &space.ModId)

	if err != nil {
		return nil, err
	}

	cachedSpaces = append(cachedSpaces, space)

	return &space, nil
}
