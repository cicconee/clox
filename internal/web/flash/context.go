package flash

import "context"

type contextKey string

var (
	messageContextKey contextKey = "flash_message"
	errorContextKey   contextKey = "flash_error"
)

func SetErrorContext(ctx context.Context, err string) context.Context {
	return context.WithValue(ctx, errorContextKey, err)
}

func GetErrorContext(ctx context.Context) string {
	flashErr, ok := ctx.Value(errorContextKey).(string)
	if !ok {
		return ""
	}

	return flashErr
}

func SetMessageContext(ctx context.Context, msg string) context.Context {
	return context.WithValue(ctx, messageContextKey, msg)
}

func GetMessageContext(ctx context.Context) string {
	flash, ok := ctx.Value(messageContextKey).(string)
	if !ok {
		return ""
	}

	return flash
}
