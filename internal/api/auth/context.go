package auth

import "context"

type contextKey string

var (
	userIDContextKey contextKey = "user_id"
)

// SetUserIDContext sets a user ID in the context. User ID can only be retrieved
// using the GetUserIDContext.
func SetUserIDContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// GetUserIDContext gets the user ID from the context.
func GetUserIDContext(ctx context.Context) string {
	userID, ok := ctx.Value(userIDContextKey).(string)
	if !ok {
		return ""
	}

	return userID
}
