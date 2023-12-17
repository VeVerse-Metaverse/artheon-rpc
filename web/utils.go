package web

import "github.com/google/uuid"

func containsUUID(arr []uuid.UUID, channel uuid.UUID) bool {
	for _, a := range arr {
		if a == channel {
			return true
		}
	}
	return false
}
