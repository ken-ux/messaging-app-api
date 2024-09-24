package defs

import "time"

type User struct {
	Username string `json:"username" validate:"required,min=5,max=20,alphanum"`
	Password string `json:"password,omitempty" validate:"required,min=5,max=20"`
}

type Message struct {
	Sender        string     `json:"sender" binding:"required"`
	Recipient     string     `json:"recipient" binding:"required"`
	Message_Body  string     `json:"message_body" binding:"required"`
	Creation_Date *time.Time `json:"creation_date,omitempty"`
}

type Profile struct {
	Description string `json:"description"`
	Color       string `json:"color,omitempty"`
}
