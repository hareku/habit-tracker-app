package auth

import (
	"context"
	"errors"
)

// UserID is the identifier of the user.
type UserID string

type authContextKey string

const userIDKey = authContextKey("auth.user-id")

// GetUserID returns the user id from the context.
func GetUserID(ctx context.Context) (UserID, bool) {
	s, ok := ctx.Value(userIDKey).(UserID)
	return s, ok
}

// MustGetUserID returns the user id from the context or panics if it's not set.
func MustGetUserID(ctx context.Context) UserID {
	s, ok := ctx.Value(userIDKey).(UserID)
	if !ok {
		panic(errors.New("missing user id in ctx"))
	}
	return s
}

// SetUserID sets the user id to the context.
func SetUserID(ctx context.Context, userID UserID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
