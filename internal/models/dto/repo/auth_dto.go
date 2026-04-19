package dto

import (
	"internship/internal/domain/entity"
	"time"
)

type UserRepo struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

func (r *UserRepo) ToEntity() *entity.User {
	u := &entity.User{
		ID:           r.ID,
		Role:         r.Role,
		Email:        r.Email,
		PasswordHash: r.PasswordHash,
		CreatedAt:    r.CreatedAt.UTC().Format(time.RFC3339),
	}

	return u
}
