package dto

type DummyLoginRequest struct {
	Role string `json:"role"` // admin, user
}

type TokenResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
type RegisterResponse struct {
	User UserData `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserData struct {
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}
