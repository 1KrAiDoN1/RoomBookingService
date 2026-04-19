package dto

import "internship/internal/domain/entity"

type CreateBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

type CreateBookingResponse struct {
	Booking BookingData `json:"booking"`
}

type ListBookingsResponse struct {
	Bookings   []BookingData   `json:"bookings"`
	Pagination PaginationQuery `json:"pagination"`
}

type CancelBookingResponse struct {
	Booking BookingData `json:"booking"`
}

type PaginationQuery struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type BookingData struct {
	BkID           string `json:"id"`
	SlotID         string `json:"slotId"`
	UserID         string `json:"userId"`
	Status         string `json:"status"`
	ConferenceLink string `json:"conferenceLink,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

func ToBookingDTO(bk *entity.Booking) BookingData {
	return BookingData{
		BkID:           bk.ID,
		SlotID:         bk.SlotID,
		UserID:         bk.UserID,
		Status:         bk.Status,
		ConferenceLink: bk.ConferenceLink,
		CreatedAt:      bk.CreatedAt,
	}
}
