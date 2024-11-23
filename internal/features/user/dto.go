package user

// Requests
type RegisterUserRequest struct {
	FirstName string `json:"firstName" validate:"required,min=2,max=15,noAllRepeatingChars"`
	LastName  string `json:"lastName" validate:"required,min=2,max=15,noAllRepeatingChars"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=5,max=10"`
}

type LoginUserRequest struct {
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required"`
	UserAgent string `json:"userAgent" validate:"required"`
	ClientIP  string `json:"clientIP" validate:"required"`
}

// Responses
type LoginUserResponse struct {
	SessionID    string `json:"sessionId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
