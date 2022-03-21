package habit

import (
	"context"
	"errors"
)

type UserID string

type authContextKey string

const userIDKey = authContextKey("user-id")

func GetUserID(ctx context.Context) (UserID, bool) {
	s, ok := ctx.Value(userIDKey).(UserID)
	return s, ok
}

func MustGetUserID(ctx context.Context) UserID {
	s, ok := ctx.Value(userIDKey).(UserID)
	if !ok {
		panic(errors.New("missing user id in ctx"))
	}
	return s
}

func SetUserID(ctx context.Context, userID UserID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
