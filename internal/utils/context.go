package utils

import (
	"context"
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
)

type ContextKey string

const UserKey = ContextKey("user")

func SetUser(r *http.Request, user *database.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserKey, user)
	return r.WithContext(ctx)
}

func GetUser(r *http.Request) *database.User {
	user, ok := r.Context().Value(UserKey).(*database.User)
	if !ok {
		return nil
	}
	return user
}
