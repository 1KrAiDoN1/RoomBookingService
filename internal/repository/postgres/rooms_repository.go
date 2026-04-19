package postgres

import (
	"context"
	"errors"
	"fmt"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	dto "internship/internal/models/dto/repo"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type roomsRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewRoomsRepository(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *roomsRepository {
	return &roomsRepository{pool: pool, getter: getter}
}

func (r *roomsRepository) CreateRoom(ctx context.Context, room *entity.Room) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	_, err := conn.Exec(ctx, `INSERT INTO rooms (id, name, description, capacity, created_at) VALUES ($1, $2, $3, $4, $5)`,
		room.ID, room.Name, room.Description, room.Capacity, room.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert room: %w", err)
	}
	return nil
}

func (r *roomsRepository) ListRooms(ctx context.Context) ([]*entity.Room, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	rows, err := conn.Query(ctx,
		`SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var result []*entity.Room
	for rows.Next() {
		row := &dto.RoomRepo{}
		if err := rows.Scan(
			&row.ID, &row.Name, &row.Description, &row.Capacity, &row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		result = append(result, row.ToEntity())
	}
	return result, rows.Err()
}

func (r *roomsRepository) GetRoomByID(ctx context.Context, id string) (*entity.Room, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	row := &dto.RoomRepo{}
	err := conn.QueryRow(ctx, `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`, id).Scan(&row.ID, &row.Name, &row.Description, &row.Capacity, &row.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorRoomNotFound
		}
		return nil, fmt.Errorf("get room by id: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *roomsRepository) CreateSchedule(ctx context.Context, s *entity.Schedule) error {

	conn := r.getter.DefaultTrOrDB(ctx, r.pool)

	_, err := conn.Exec(ctx, `INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES ($1, $2, $3, $4, $5)`,
		s.ID, s.RoomID, s.DaysOfWeek, s.StartTime, s.EndTime,
	)
	if err != nil {
		return fmt.Errorf("insert schedule: %w", err)
	}
	return nil
}

func (r *roomsRepository) GetScheduleByRoomID(ctx context.Context, roomID string) (*entity.Schedule, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	row := &dto.ScheduleRepo{}
	err := conn.QueryRow(ctx, `SELECT id, room_id, days_of_week, start_time, end_time FROM schedules WHERE room_id = $1`, roomID).Scan(&row.ID, &row.RoomID, &row.DaysOfWeek, &row.StartTime, &row.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *roomsRepository) IsExistsSchedule(ctx context.Context, roomID string) (bool, error) {
	var exists bool

	conn := r.getter.DefaultTrOrDB(ctx, r.pool)

	err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schedules WHERE room_id = $1)`, roomID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check schedule exists: %w", err)
	}
	return exists, nil
}

func (r *roomsRepository) GetAvailableSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error) {
	dayStart, dayEnd, err := dayBounds(date)
	if err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	rows, err := conn.Query(ctx,
		`SELECT s.id, s.room_id, s.start_at, s.end_at FROM slots s
		 LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
		 WHERE s.room_id = $1 AND s.start_at >= $2 AND s.start_at <  $3 AND b.id IS NULL ORDER BY s.start_at`,
		roomID, dayStart, dayEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("get available slots: %w", err)
	}
	defer rows.Close()

	return scanSlots(rows)
}

func (r *roomsRepository) GetBookedSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error) {
	dayStart, dayEnd, err := dayBounds(date)
	if err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	rows, err := conn.Query(ctx,
		`SELECT s.id, s.room_id, s.start_at, s.end_at FROM slots s INNER JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
		 WHERE s.room_id = $1 AND s.start_at >= $2 AND s.start_at <  $3 ORDER BY s.start_at`,
		roomID, dayStart, dayEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("get booked slots: %w", err)
	}
	defer rows.Close()

	return scanSlots(rows)
}

func (r *roomsRepository) GetSlotByID(ctx context.Context, id string) (*entity.Slot, error) {
	row := &dto.SlotRepo{}
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	err := conn.QueryRow(ctx, `SELECT id, room_id, start_at, end_at FROM slots WHERE id = $1`, id).Scan(&row.ID, &row.RoomID, &row.Start, &row.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorSlotNotFound
		}
		return nil, fmt.Errorf("get slot by id: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *roomsRepository) IsAvailableSlot(ctx context.Context, slotID string) (bool, error) {
	var exists bool

	conn := r.getter.DefaultTrOrDB(ctx, r.pool)

	err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings WHERE slot_id = $1 AND status = 'active')`, slotID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check slot availability: %w", err)
	}
	return !exists, nil
}

func (r *roomsRepository) UpsertSlots(ctx context.Context, slots []*entity.Slot) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	for _, s := range slots {
		startTime, err := time.Parse(time.RFC3339, s.Start)
		if err != nil {
			return fmt.Errorf("parse slot start %q: %w", s.Start, err)
		}
		endTime, err := time.Parse(time.RFC3339, s.End)
		if err != nil {
			return fmt.Errorf("parse slot end %q: %w", s.End, err)
		}
		_, err = conn.Exec(ctx,
			`INSERT INTO slots (id, room_id, start_at, end_at)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (room_id, start_at) DO NOTHING`,
			s.ID, s.RoomID, startTime, endTime,
		)
		if err != nil {
			return fmt.Errorf("upsert slot: %w", err)
		}
	}
	return nil
}

func dayBounds(date string) (time.Time, time.Time, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 0, 1), nil
}

func scanSlots(rows pgx.Rows) ([]*entity.Slot, error) {
	var result []*entity.Slot
	for rows.Next() {
		row := &dto.SlotRepo{}
		if err := rows.Scan(
			&row.ID, &row.RoomID, &row.Start, &row.End,
		); err != nil {
			return nil, fmt.Errorf("scan slot: %w", err)
		}
		result = append(result, row.ToEntity())
	}
	return result, rows.Err()
}
