package models

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type Presence struct {
	Status   string    `json:"status"`
	UserId   uuid.UUID `json:"userId,omitempty"`
	SpaceId  uuid.UUID `json:"spaceId,omitempty"`
	ServerId uuid.UUID `json:"serverId,omitempty"`
}

func (presence *Presence) Reset() {
	presence.UserId = uuid.Nil
	presence.ServerId = uuid.Nil
	presence.SpaceId = uuid.Nil
	presence.Status = "offline"
}

func UpdateUserPresenceStatus(db *sql.DB, id uuid.UUID, inPresence Presence) (*Presence, error) {
	user, err := GetUserById(db, id)
	if err != nil {
		return nil, err
	}

	presence := Presence{}

	presence.Reset()

	// try to find existing record
	err = db.QueryRow(
		"SELECT p.user_id, p.status, p.space_id, p.server_id FROM presence p WHERE p.user_id = $1",
		user.Id,
	).Scan(&presence.UserId, &presence.Status, &presence.SpaceId, &presence.ServerId)

	var spaceId *uuid.UUID
	if inPresence.SpaceId != uuid.Nil {
		spaceId = &inPresence.SpaceId
	}

	var serverId *uuid.UUID
	if inPresence.ServerId != uuid.Nil {
		serverId = &inPresence.ServerId
	}

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		// insert a new presence record for the user
		rows, err := db.Query(
			"INSERT INTO presence (user_id, status, space_id, server_id) VALUES ($1, $2, $3, $4)",
			user.Id,
			inPresence.Status,
			spaceId,
			serverId,
		)

		if err != nil {
			return nil, err
		}

		err = rows.Close()
		if err != nil {
			fmt.Printf("failed to close errors: %s", err.Error())
		}

	} else {
		// update an existing presence record for the user
		rows, err := db.Query(
			"UPDATE presence SET status=$2, space_id=$3, server_id=$4 WHERE user_id=$1",
			user.Id,
			inPresence.Status,
			spaceId,
			serverId,
		)

		if err != nil {
			return nil, err
		}

		err = rows.Close()
		if err != nil {
			fmt.Printf("failed to close errors: %s", err.Error())
		}
	}

	presence.Reset()

	err = db.QueryRow(
		"SELECT p.user_id, p.status, p.space_id, p.server_id FROM presence p WHERE p.user_id = $1",
		user.Id,
	).Scan(&presence.UserId, &presence.Status, &presence.SpaceId, &presence.ServerId)

	if err != nil {
		return nil, err
	}

	return &presence, nil
}
