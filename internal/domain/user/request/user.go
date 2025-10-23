package request

// RegisterUser represents a user registration request
type RegisterUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,len=10"`
}

// LoginUser represents a user login request
type LoginUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateUser represents a user update request
type UpdateUser struct {
	FullName string `json:"full_name" validate:"min=2,max=255"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,len=10"`
}
