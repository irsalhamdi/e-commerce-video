package course

import "time"

type Course struct {
	ID          string    `json:"id" db:"course_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ImageURL    string    `json:"imageUrl" db:"image_url"`
	Price       int       `json:"price" db:"price"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	Version     int       `json:"-" db:"version"`
}

type CourseNew struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Price       int    `json:"price" validate:"required,gte=0,lte=10000"`
	ImageURL    string `json:"imageUrl" validate:"required"`
}

type CourseUp struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int    `json:"price" validate:"omitempty,gte=0,lte=10000"`
	ImageURL    *string `json:"imageUrl"`
}
