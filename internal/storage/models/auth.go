package models

type AuthenticateRequest struct {
	Token string
	Roles []string
}

type User struct {
	Id         string   `json:"id"`
	LastName   *string  `json:"lastName"`
	FirstName  *string  `json:"firstName"`
	MiddleName *string  `json:"middleName"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
}
