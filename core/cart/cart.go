package cart

import (
	"time"
)

type Cart struct {
	UserID    string    `json:"-" db:"user_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int       `json:"-" db:"version"`
	Items     []Item    `json:"items" db:"-"`
}

type Item struct {
	UserID    string    `json:"-" db:"user_id"`
	CourseID  string    `json:"courseId" db:"course_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type ItemNew struct {
	CourseID string `json:"courseId" db:"course_id"`
}
