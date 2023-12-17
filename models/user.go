package models

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Presence Presence  `json:"-"`
}

var cachedUsers = make([]User, 0)

func (user *User) UpdateUserPresence(db *sql.DB, status string, spaceId uuid.UUID, serverId uuid.UUID) (err error) {
	user.Presence.Status = status
	user.Presence.SpaceId = spaceId
	user.Presence.ServerId = serverId

	presence, err := UpdateUserPresenceStatus(db, user.Id, user.Presence)
	if err != nil {
		return err
	}

	if presence == nil {
		return fmt.Errorf("no presence record")
	}

	user.Presence = *presence

	return nil
}

func GetCachedUserById(id uuid.UUID) *User {
	for _, v := range cachedUsers {
		if v.Id == id {
			return &v
		}
	}
	return nil
}

func GetUserById(db *sql.DB, id uuid.UUID) (*User, error) {
	//region cache
	cachedUser := GetCachedUserById(id)
	if cachedUser != nil {
		return cachedUser, nil
	}
	//endregion cache

	//region database
	user := User{}

	err := db.QueryRow(
		"SELECT e.id, u.name FROM users AS u LEFT JOIN entities AS e ON u.id=e.id where u.id = $1",
		id,
	).Scan(&user.Id, &user.Name)

	if err != nil {
		return nil, err
	}
	//endregion database

	cachedUsers = append(cachedUsers, user) // add to the cache

	return &user, nil
}

var cachedLeaderMap = make(map[uuid.UUID][]User)

func GetCachedLeadersByUserId(userId uuid.UUID) []User {
	for k, v := range cachedLeaderMap {
		if k == userId {
			return v
		}
	}
	return nil
}

func GetUserLeadersById(db *sql.DB, userId uuid.UUID) ([]User, error) {

	//region cache
	cachedLeaders := GetCachedLeadersByUserId(userId)
	if cachedLeaders != nil {
		return cachedLeaders, nil
	}
	//endregion cache

	//region database
	rows, err := db.Query(
		`SELECT f.leader_id AS id, u.name AS name 
FROM followers AS f
LEFT JOIN users AS u ON f.leader_id = u.id
WHERE (f.follower_id = $1) `, //AND f.leader_id != f.follower_id
		userId,
	)

	leaders := make([]User, 0)

	if err != nil {
		if err == sql.ErrNoRows {
			return leaders, nil
		} else {
			return leaders, err
		}
	}

	for rows.Next() {
		var leader User
		if err := rows.Scan(&leader.Id, &leader.Name); err != nil {
			return leaders, err
		} else {
			leaders = append(leaders, leader)
		}
	}
	//endregion database

	cachedLeaderMap[userId] = leaders

	return leaders, nil
}
