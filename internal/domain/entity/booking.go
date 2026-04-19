package entity

type Booking struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      string
}

const (
	BookingStatusActive    string = "active"
	BookingStatusCancelled string = "cancelled"
)
