package models

import "time"

// User represents a system user
type User struct {
	ID        int       `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"` // Password is never sent to the client
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}
