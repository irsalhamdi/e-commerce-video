package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/irsalhamdi/e-commerce-video/api/web"
	"github.com/irsalhamdi/e-commerce-video/api/weberr"
	"github.com/irsalhamdi/e-commerce-video/core/claims"
	"github.com/irsalhamdi/e-commerce-video/validate"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func HandleCreate(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		var u UserNew
		if err := web.Decode(w, r, &u); err != nil {
			return weberr.BadRequest(fmt.Errorf("unable to decode payload: %w", err))
		}

		if !claims.IsAdmin(ctx) {
			return weberr.NotAuthorized(errors.New("only admin can create other admins"))
		}

		if err := validate.Check(u); err != nil {
			return weberr.NewError(err, err.Error(), http.StatusUnprocessableEntity)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating password hash: %w", err)
		}

		now := time.Now().UTC()

		usr := User{
			ID:           validate.GenerateID(),
			Name:         u.Name,
			Email:        u.Email,
			Role:         u.Role,
			PasswordHash: hash,
			CreatedAt:    now,
			UpdatedAt:    now,
			Active:       true,
		}

		if err := Create(ctx, db, usr); err != nil {
			err := fmt.Errorf("creating user[%s]: %w", u.Email, err)
			if errors.Is(err, ErrUniqueEmail) {
				return weberr.NewError(err, ErrUniqueEmail.Error(), http.StatusConflict)
			}
			return err
		}

		return web.Respond(ctx, w, usr, http.StatusCreated)
	}
}

func HandleShow(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		userID := web.Param(r, "id")
		if err := validate.CheckID(userID); err != nil {
			return weberr.NewError(err, err.Error(), http.StatusUnprocessableEntity)
		}

		if !claims.IsUser(ctx, userID) && !claims.IsAdmin(ctx) {
			return weberr.NotAuthorized(errors.New("user trying to fetch another user"))
		}

		user, err := Fetch(ctx, db, userID)
		if err != nil {
			return fmt.Errorf("fetching user[%s]: %w", userID, err)
		}

		return web.Respond(ctx, w, user, http.StatusOK)
	}
}

func HandleShowCurrent(db *sqlx.DB) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		clm, err := claims.Get(ctx)
		if err != nil {
			return weberr.NotAuthorized(errors.New("user not authenticated"))
		}

		user, err := Fetch(ctx, db, clm.UserID)
		if err != nil {
			return fmt.Errorf("fetching user[%s]: %w", clm.UserID, err)
		}

		return web.Respond(ctx, w, user, http.StatusOK)
	}
}
