package session

import "context"

type contextKey string

var userContextKey contextKey = "user_session"

func SetUserContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUserContext(ctx context.Context) User {
	user, ok := ctx.Value(userContextKey).(User)
	if !ok {
		return User{}
	}

	return user
}
