package models

import "github.com/google/uuid"

type LeaderboardEntry struct {
	UserID   uuid.UUID
	Login    string
	WinRatio float64
}
