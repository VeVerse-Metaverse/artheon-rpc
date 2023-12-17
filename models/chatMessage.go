package models

import (
	"database/sql"
	"github.com/google/uuid"
)

type ChatMessage struct {
	Id              string    `json:"id,omitempty"`
	UserId          uuid.UUID `json:"userId"`
	Message         string    `json:"message"`
	ChannelId       string    `json:"channelId"`
	ChannelName     string    `json:"channelName"`
	ChannelCategory string    `json:"channelCategory"`
}

func AddChatMessage(db *sql.DB, m ChatMessage) error {
	query := "INSERT INTO chat_messages (user_id, message, channel_id, channel_name, channel_category) VALUES ($1, $2, $3, $4, $5)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		m.UserId,
		m.Message,
		m.ChannelId,
		m.ChannelName,
		m.ChannelCategory)

	return err
}
