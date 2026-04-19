package dto

import (
	"internship/internal/domain/entity"
	"time"
)

type BookingRepo struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      time.Time
}

func (r *BookingRepo) ToEntity() *entity.Booking {
	return &entity.Booking{
		ID:             r.ID,
		SlotID:         r.SlotID,
		UserID:         r.UserID,
		Status:         r.Status,
		ConferenceLink: r.ConferenceLink,
		CreatedAt:      r.CreatedAt.UTC().Format(time.RFC3339),
	}
}
