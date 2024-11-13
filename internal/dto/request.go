package dto

type RegisterUserRequest struct {
	FirstName string `json:"firstName" validate:"required,min=2,max=15,noAllRepeatingChars"`
	LastName  string `json:"lastName" validate:"required,min=2,max=15,noAllRepeatingChars"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=5,max=10"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}
