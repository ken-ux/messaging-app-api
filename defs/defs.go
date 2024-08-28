package defs

type User struct {
	Username string `json:"username" validate:"required,min=5,max=20,alphanum"`
	Password string `json:"password" validate:"required,min=5,max=20"`
}
