package server

import (
	"context"
)

type AuthHandlerInterface interface {
	MergeAuthKey(ctx context.Context, existingAuth, newAuth string) (string, error)

	ValidateAuthKey(ctx context.Context, authKey string) error
}

type AuthHandlerProvider interface {
	AuthHandlerInterface
}
