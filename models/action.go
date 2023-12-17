package models

import "database/sql"

type Action struct {
	Id       string `json:"id,omitempty"`
	UserId   string `json:"userId"`
	SenderId string `json:"senderId"`
	Details  string `json:"details,omitempty"`
	Action   string `json:"action"`
}

func AddAction(db *sql.DB, action Action) error {
	query := "INSERT INTO actions (sender_id, user_id, details, action) VALUES ($1, $2, $3, $4)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		action.SenderId,
		action.UserId,
		action.Details,
		action.Action)

	return err
}

func GetActionById(db *sql.DB, actionId string) (*Action, error) {

	action := Action{}

	err := db.QueryRow(
		"SELECT id, action, details, user_id, sender_id FROM actions AS a WHERE id = $1",
		actionId,
	).Scan(&action.Id, &action.Action, &action.Details, &action.UserId, &action.SenderId)

	if err != nil {
		return nil, err
	}

	return &action, nil
}
