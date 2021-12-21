package context

import "context"

type contextKey string

func (c contextKey) String() string {
	return "app context key " + string(c)
}

const (
	userIDKey = contextKey("user_id")
)

// GetUserID reads the user ID from the context.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

// WithUserID adds the user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
