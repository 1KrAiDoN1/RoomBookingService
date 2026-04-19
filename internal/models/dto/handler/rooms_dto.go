package dto

import "internship/internal/domain/entity"

type CreateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

type CreateRoomResponse struct {
	Room RoomData `json:"room"`
}

type ListRoomsResponse struct {
	Rooms []RoomData `json:"rooms"`
}

type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CreateScheduleResponse struct {
	Schedule ScheduleData `json:"schedule"`
}

type ListSlotsResponse struct {
	Slots []SlotData `json:"slots"`
}

type RoomData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Capacity    int    `json:"capacity,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

type ScheduleData struct {
	ID         string `json:"id"`
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type SlotData struct {
	ID     string `json:"id"`
	RoomID string `json:"roomId"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

func ToRoomDTO(room *entity.Room) RoomData {
	return RoomData{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
		CreatedAt:   room.CreatedAt,
	}
}

func ToScheduleDTO(s *entity.Schedule) ScheduleData {
	return ScheduleData{
		ID:         s.ID,
		RoomID:     s.RoomID,
		DaysOfWeek: s.DaysOfWeek,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
	}
}

func ToSlotDTO(s *entity.Slot) SlotData {
	return SlotData{
		ID:     s.ID,
		RoomID: s.RoomID,
		Start:  s.Start,
		End:    s.End,
	}
}
