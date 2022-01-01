package context

import (
	"context"
	"mmr/models"
)

type contextKey string

func (c contextKey) String() string {
	return "app context key " + string(c)
}

const (
	uuidKey   = contextKey("uuid")
	userIDKey = contextKey("user_id")
	userKey   = contextKey("user")
)

func GetUserID(ctx context.Context) int32 {
	id, _ := ctx.Value(userIDKey).(int32)
	return id
}

func WithUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUser(ctx context.Context) *models.User {
	usr, _ := ctx.Value(userKey).(models.User)
	return &usr
}

func WithUser(ctx context.Context, usr *models.User) context.Context {
	return context.WithValue(ctx, userKey, *usr)
}

func GetUUID(ctx context.Context) string {
	id, _ := ctx.Value(uuidKey).(string)
	return id
}

func WithUUID(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, uuidKey, uuid)
}
