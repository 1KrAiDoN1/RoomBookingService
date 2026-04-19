package dto

import (
	"internship/internal/domain/entity"
	"time"
)

type RoomRepo struct {
	ID          string
	Name        string
	Description string
	Capacity    int
	CreatedAt   time.Time
}

func (r *RoomRepo) ToEntity() *entity.Room {
	return &entity.Room{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Capacity:    r.Capacity,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type ScheduleRepo struct {
	ID         string
	RoomID     string
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

func (r *ScheduleRepo) ToEntity() *entity.Schedule {

	return &entity.Schedule{
		ID:         r.ID,
		RoomID:     r.RoomID,
		DaysOfWeek: r.DaysOfWeek,
		StartTime:  r.StartTime,
		EndTime:    r.EndTime,
	}
}

type SlotRepo struct {
	ID     string
	RoomID string
	Start  time.Time
	End    time.Time
}

func (r *SlotRepo) ToEntity() *entity.Slot {
	return &entity.Slot{
		ID:     r.ID,
		RoomID: r.RoomID,
		Start:  r.Start.UTC().Format(time.RFC3339),
		End:    r.End.UTC().Format(time.RFC3339),
	}
}
