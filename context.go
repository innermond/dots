package dots

import (
	"context"
	"log"
)

type key int

const (
	userContextKey = key(iota + 1)
	flashContextKey
)

func NewContextWithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func UserFromContext(ctx context.Context) *User {
	u, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		log.Printf("dots: user not found in context: %v\n", u)
		return &User{}
	}

	return u
}
