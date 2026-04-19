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

type bookingRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewBookingRepository(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *bookingRepository {
	return &bookingRepository{pool: pool, getter: getter}
}

func (r *bookingRepository) CreateBooking(ctx context.Context, booking *entity.Booking) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)

	createdAt, err := time.Parse(time.RFC3339, booking.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}

	_, err = conn.Exec(ctx, `INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink, createdAt)
	if err != nil {
		return fmt.Errorf("insert booking: %w", err)
	}

	return nil
}

func (r *bookingRepository) GetBookingByID(ctx context.Context, id string) (*entity.Booking, error) {
	row := &dto.BookingRepo{}
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	err := conn.QueryRow(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings WHERE id = $1`, id,
	).Scan(
		&row.ID, &row.SlotID, &row.UserID, &row.Status, &row.ConferenceLink, &row.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorBookingNotFound
		}
		return nil, fmt.Errorf("get booking by id: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *bookingRepository) ListAllBookings(ctx context.Context, page, pageSize int) ([]*entity.Booking, int, error) {
	offset := (page - 1) * pageSize
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)

	rows, err := conn.Query(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at, COUNT(*) OVER() AS total FROM bookings ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list all bookings: %w", err)
	}
	defer rows.Close()

	var (
		result []*entity.Booking
		total  int
	)
	for rows.Next() {
		row := &dto.BookingRepo{}
		if err := rows.Scan(
			&row.ID, &row.SlotID, &row.UserID, &row.Status, &row.ConferenceLink, &row.CreatedAt,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("scan booking: %w", err)
		}
		result = append(result, row.ToEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (r *bookingRepository) ListMyBookings(ctx context.Context, userID string, nowStr time.Time) ([]*entity.Booking, error) {

	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	rows, err := conn.Query(ctx,
		`SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
		 FROM bookings b
		 INNER JOIN slots s ON s.id = b.slot_id
		 WHERE b.user_id = $1
		   AND s.start_at >= $2
		 ORDER BY s.start_at`,
		userID, nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("list my bookings: %w", err)
	}
	defer rows.Close()

	var result []*entity.Booking
	for rows.Next() {
		row := &dto.BookingRepo{}
		if err := rows.Scan(
			&row.ID, &row.SlotID, &row.UserID, &row.Status, &row.ConferenceLink, &row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		result = append(result, row.ToEntity())
	}
	return result, rows.Err()
}

func (r *bookingRepository) CancelBooking(ctx context.Context, id string) (*entity.Booking, error) {
	row := &dto.BookingRepo{}
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	err := conn.QueryRow(ctx,
		`UPDATE bookings
		 SET status = 'cancelled'
		 WHERE id = $1
		 RETURNING id, slot_id, user_id, status, conference_link, created_at`, id,
	).Scan(
		&row.ID, &row.SlotID, &row.UserID, &row.Status, &row.ConferenceLink, &row.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorBookingNotFound
		}
		return nil, fmt.Errorf("cancel booking: %w", err)
	}
	return row.ToEntity(), nil
}
