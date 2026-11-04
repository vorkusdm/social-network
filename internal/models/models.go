package models
import "time"

type User struct {
	ID         int       `json:"id"`
	Login      string    `json:"login"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	BirthDate  time.Time `json:"birth_date"`
	Gender     string    `json:"gender"`
	Interests  string    `json:"interests"`
	City       string    `json:"city"`
	Password   string    `json:"password,omitempty"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}