package userctx

import (
	"context"

	"github.com/linkasu/linka.type-backend/internal/auth"
)

type ctxKey struct{}

// With stores auth user in context.
func With(ctx context.Context, user auth.User) context.Context {
	return context.WithValue(ctx, ctxKey{}, user)
}

// From extracts auth user from context.
func From(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(ctxKey{}).(auth.User)
	return user, ok
}
