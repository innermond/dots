package dots

import "context"

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
		// TODO reporting not ok
		return UserZero
	}

	return u
}
