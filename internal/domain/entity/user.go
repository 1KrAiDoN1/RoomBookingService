package entity

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    string
}

const (
	RoleAdmin string = "admin"
	RoleUser  string = "user"
)
