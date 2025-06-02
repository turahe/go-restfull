package requests

import "github.com/google/uuid"

type AuthLoginRequest struct {
	UserName string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password"`
}

type AuthRegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	UserName        string `json:"username" validate:"required,min=3,max=32"`
	Phone           string `json:"phone" validate:"required,numeric"`
	Password        string `json:"password" validate:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type CreateUserRequest struct {
	UserName string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}
type ResetPasswordRequest struct {
	Password        string `json:"password" validate:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" validate:"required,min=8,max=32"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
type ChangePhoneRequest struct {
	Phone string `json:"phone" validate:"required,numeric"`
}
type ChangeEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}
type ChangeUserNameRequest struct {
	UserName string `json:"username" validate:"required,min=3,max=32"`
}

type UpdateUserRequest struct {
	UserName string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,numeric"`
}

type GetUserNameRequest struct {
	UserName string `json:"username" validate:"required,min=3,max=32"`
}

type GetUserEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}
type GetUserPhoneRequest struct {
	Phone string `json:"phone" validate:"required,numeric"`
}
type GetUserIdRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
}

type ChangeAvatarRequest struct {
	Image []byte `json:"image" validate:"required,base64"`
}
