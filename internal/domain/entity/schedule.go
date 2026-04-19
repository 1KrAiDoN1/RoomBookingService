package entity

type Schedule struct {
	ID         string
	RoomID     string
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}
