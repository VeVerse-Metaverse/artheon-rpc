package models

import (
	"database/sql"
	"github.com/google/uuid"
)

type Server struct {
	Id      uuid.UUID `json:"id"`
	Host    string    `json:"host,omitempty"`
	Port    int       `json:"port,omitempty"`
	SpaceId uuid.UUID `json:"spaceId,omitempty"`
	Public  bool      `json:"public"`
}

var cachedServers = make([]Server, 0)

func GetCachedServerById(id uuid.UUID) *Server {
	for _, v := range cachedServers {
		if v.Id == id {
			return &v
		}
	}
	return nil
}

func GetServerById(db *sql.DB, serverId uuid.UUID) (*Server, error) {

	//region cache
	cachedServer := GetCachedServerById(serverId)
	if cachedServer != nil {
		return cachedServer, nil
	}
	//endregion

	server := Server{}

	err := db.QueryRow(
		"SELECT s.id, s.host, s.port, s.space_id, s.public FROM servers AS s WHERE s.id = $1",
		serverId,
	).Scan(&server.Id, &server.Host, &server.Port, &server.SpaceId, &server.Public)

	if err != nil {
		return nil, err
	}

	cachedServers = append(cachedServers, server) // add to the cache

	return &server, nil
}
