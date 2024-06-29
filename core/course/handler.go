package course

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/irsalhamdi/e-commerce-video/api/web"
	"github.com/irsalhamdi/e-commerce-video/api/weberr"
	"github.com/irsalhamdi/e-commerce-video/core/claims"
	"github.com/irsalhamdi/e-commerce-video/database"
	"github.com/irsalhamdi/e-commerce-video/validate"
	"github.com/jmoiron/sqlx"
)

func HandleCreate(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		var c CourseNew
		if err := web.Decode(w, r, &c); err != nil {
			return weberr.BadRequest(fmt.Errorf("unable to decode payload: %w", err))
		}

		if err := validate.Check(c); err != nil {
			return weberr.NewError(err, err.Error(), http.StatusUnprocessableEntity)
		}

		now := time.Now().UTC()

		course := Course{
			ID:          validate.GenerateID(),
			Name:        c.Name,
			Description: c.Description,
			Price:       c.Price,
			ImageURL:    c.ImageURL,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := Create(ctx, db, course); err != nil {
			if errors.Is(err, database.ErrDBDuplicatedEntry) {
				return weberr.NewError(err, "passed course already exists", http.StatusUnprocessableEntity)
			}
			return err
		}

		return web.Respond(ctx, w, course, http.StatusCreated)
	}
}

func HandleUpdate(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		courseID := web.Param(r, "id")

		if err := validate.CheckID(courseID); err != nil {
			return weberr.NewError(err, err.Error(), http.StatusUnprocessableEntity)
		}

		var cup CourseUp
		if err := web.Decode(w, r, &cup); err != nil {
			return weberr.BadRequest(fmt.Errorf("unable to decode payload: %w", err))
		}

		if err := validate.Check(cup); err != nil {
			return weberr.NewError(err, err.Error(), http.StatusUnprocessableEntity)
		}

		course, err := Fetch(ctx, db, courseID)
		if err != nil {
			err := fmt.Errorf("fetching passed course[%s]: %w", courseID, err)
			if errors.Is(err, database.ErrDBNotFound) {
				return weberr.NotFound(err)
			}
			return err
		}

		if cup.Name != nil {
			course.Name = *cup.Name
		}
		if cup.Description != nil {
			course.Description = *cup.Description
		}
		if cup.Price != nil {
			course.Price = *cup.Price
		}
		if cup.ImageURL != nil {
			course.ImageURL = *cup.ImageURL
		}
		course.UpdatedAt = time.Now().UTC()

		if course, err = Update(ctx, db, course); err != nil {
			return fmt.Errorf("updating course[%s]: %w", course.ID, err)
		}

		return web.Respond(ctx, w, course, http.StatusOK)
	}
}

func HandleList(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		courses, err := FetchAll(ctx, db)
		if err != nil {
			return fmt.Errorf("fetching all courses: %w", err)
		}

		return web.Respond(ctx, w, courses, http.StatusOK)
	}
}

func HandleListOwned(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		clm, err := claims.Get(ctx)
		if err != nil {
			return weberr.NotAuthorized(errors.New("user not authenticated"))
		}

		courses, err := FetchByOwner(ctx, db, clm.UserID)
		if err != nil {
			return fmt.Errorf("fetching courses of user[%s]: %w", clm.UserID, err)
		}

		return web.Respond(ctx, w, courses, http.StatusOK)
	}
}

func HandleShow(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		courseID := web.Param(r, "id")

		if err := validate.CheckID(courseID); err != nil {
			return weberr.NewError(err, err.Error(), http.StatusUnprocessableEntity)
		}

		course, err := Fetch(ctx, db, courseID)
		if err != nil {
			err := fmt.Errorf("fetching course[%s]: %w", courseID, err)
			if errors.Is(err, database.ErrDBNotFound) {
				return weberr.NotFound(err)
			}
			return err
		}

		return web.Respond(ctx, w, course, http.StatusOK)
	}
}