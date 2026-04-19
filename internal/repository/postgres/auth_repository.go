package postgres

import (
	"context"
	"errors"
	"fmt"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	dto "internship/internal/models/dto/repo"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewAuthRepository(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *authRepository {
	return &authRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *authRepository) CreateUser(ctx context.Context, user *entity.User) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	_, err := conn.Exec(ctx, `INSERT INTO users (id, email, password_hash, role, created_at) VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	row := &dto.UserRepo{}
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	err := conn.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.Role, &row.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	row := &dto.UserRepo{}
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	err := conn.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.Role, &row.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return row.ToEntity(), nil
}
