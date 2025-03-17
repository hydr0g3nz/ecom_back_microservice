package dto

import "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"

type UserRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

func (u UserRequest) ToEntity() entity.User {
	return entity.User{
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type TokenRequest struct {
	Token string `json:"token" validate:"required"`
}
