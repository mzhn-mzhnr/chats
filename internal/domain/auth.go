package domain

type AuthRequest struct {
	Token string
	Roles []string
}
