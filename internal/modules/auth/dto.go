package auth

import "time"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`

	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Email *string `json:"email" binding:"omitempty,email"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

type AdminResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

type LinkIdentityRequest struct {
	Provider   string `json:"provider" binding:"required"`
	Identifier string `json:"identifier" binding:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
}

type UserResponse struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type MeResponse struct {
	User        *UserResponse `json:"user"`
	AccessToken string        `json:"accessToken,omitempty"`
}
