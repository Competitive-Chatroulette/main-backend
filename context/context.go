package context

import (
	"context"
	"github.com/golang-jwt/jwt"
)

type contextKey string

func (c contextKey) String() string {
	return "app context key " + string(c)
}

const (
	jwtKey    = contextKey("jwt")
	uuidKey   = contextKey("uuid")
	userIDKey = contextKey("user_id")
)

func GetUserID(ctx context.Context) int32 {
	id, _ := ctx.Value(userIDKey).(int32)
	return id
}

func WithUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetJwt(ctx context.Context) *jwt.Token {
	jwt, _ := ctx.Value(jwtKey).(*jwt.Token)
	return jwt
}

func WithJwt(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, jwtKey, token)
}

func GetUUID(ctx context.Context) string {
	id, _ := ctx.Value(uuidKey).(string)
	return id
}

func WithUUID(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, uuidKey, uuid)
}
