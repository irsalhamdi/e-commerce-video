package claims

import (
	"context"
	"errors"
)

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

type Claims struct {
	UserID string
	Role   string
}

type ctxKey int

const claimsKey ctxKey = 1

func Set(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func Get(ctx context.Context) (Claims, error) {
	v, ok := ctx.Value(claimsKey).(Claims)
	if !ok {
		return Claims{}, errors.New("claim value missing from context")
	}
	return v, nil
}

func IsAdmin(ctx context.Context) bool {
	c, err := Get(ctx)
	if err != nil {
		return false
	}

	return c.Role == RoleAdmin
}

func IsUser(ctx context.Context, id string) bool {
	c, err := Get(ctx)
	if err != nil {
		return false
	}

	return c.UserID == id
}
