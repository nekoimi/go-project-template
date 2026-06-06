package auth

const EventUserRegistered = "user.registered"

type UserRegisteredEvent struct {
	UserID   string
	Username string
	Email    string
}
