package service

import (
	"context"
	"fmt"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const slotDuration = 30 * time.Minute

type roomsService struct {
	roomsRepository RoomsRepositoryInterface
	trManager       TransactionManagerInterface
	logger          *zap.Logger
}

func NewRoomsService(roomsRepository RoomsRepositoryInterface, trManager TransactionManagerInterface, logger *zap.Logger) *roomsService {
	return &roomsService{
		roomsRepository: roomsRepository,
		trManager:       trManager,
		logger:          logger,
	}
}

func (s *roomsService) CreateRoom(ctx context.Context, name, description string, capacity int) (*entity.Room, error) {
	room := &entity.Room{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if err := s.roomsRepository.CreateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return room, nil
}

func (s *roomsService) ListRooms(ctx context.Context) ([]*entity.Room, error) {
	rooms, err := s.roomsRepository.ListRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	if rooms == nil {
		rooms = []*entity.Room{}
	}
	return rooms, nil
}

func (s *roomsService) CreateSchedule(ctx context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*entity.Schedule, error) {
	room, err := s.roomsRepository.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	if room == nil {
		return nil, domain.ErrorRoomNotFound
	}

	for _, d := range daysOfWeek {
		if d < 1 || d > 7 {
			return nil, domain.ErrorInvalidRequest
		}
	}

	if !validHHMM(startTime) || !validHHMM(endTime) || startTime >= endTime {
		return nil, domain.ErrorInvalidRequest
	}

	schedule := &entity.Schedule{
		ID:         uuid.New().String(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	err = s.trManager.Do(ctx, func(txCtx context.Context) error {
		exist, err := s.roomsRepository.IsExistsSchedule(txCtx, roomID)
		if err != nil {
			return fmt.Errorf("check schedule exists: %w", err)
		}
		if exist {
			return domain.ErrorScheduleExists
		}

		if err = s.roomsRepository.CreateSchedule(txCtx, schedule); err != nil {
			return fmt.Errorf("create schedule: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *roomsService) ListAvailableSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, domain.ErrorInvalidRequest
	}

	room, err := s.roomsRepository.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	if room == nil {
		return nil, domain.ErrorRoomNotFound
	}

	schedule, err := s.roomsRepository.GetScheduleByRoomID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}

	if schedule == nil {
		return []*entity.Slot{}, nil
	}

	if !containsDay(schedule.DaysOfWeek, isoWeekday(d)) {
		return []*entity.Slot{}, nil
	}

	generated := generateSlots(roomID, d, schedule)

	if len(generated) > 0 {
		if err := s.roomsRepository.UpsertSlots(ctx, generated); err != nil {
			s.logger.Error("failed to upsert slots", zap.Error(err))
		}
	}

	availableSlots, err := s.roomsRepository.GetAvailableSlots(ctx, roomID, date)
	if err != nil {
		return nil, fmt.Errorf("get available slots: %w", err)
	}

	return availableSlots, nil
}

func generateSlots(roomID string, date time.Time, sch *entity.Schedule) []*entity.Slot {
	y, m, d := date.UTC().Date()

	start, err := parseHHMM(y, m, d, sch.StartTime)
	if err != nil {
		return nil
	}
	end, err := parseHHMM(y, m, d, sch.EndTime)
	if err != nil {
		return nil
	}

	now := time.Now().UTC()
	var slots []*entity.Slot

	for cur := start; !cur.Add(slotDuration).After(end); cur = cur.Add(slotDuration) {
		slotEnd := cur.Add(slotDuration)
		if slotEnd.Before(now) {
			continue
		}
		slots = append(slots, &entity.Slot{
			ID:     uuid.New().String(),
			RoomID: roomID,
			Start:  cur.UTC().Format(time.RFC3339),
			End:    slotEnd.UTC().Format(time.RFC3339),
		})
	}
	return slots
}

func parseHHMM(y int, m time.Month, d int, hhmm string) (time.Time, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(y, m, d, t.Hour(), t.Minute(), 0, 0, time.UTC), nil
}

func isoWeekday(t time.Time) int {
	wd := int(t.UTC().Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func containsDay(days []int, day int) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

func validHHMM(s string) bool {
	_, err := time.Parse("15:04", s)
	return err == nil
}
