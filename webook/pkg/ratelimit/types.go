package ratelimit

import "context"

//go:generate mockgen -source=./ratelimit/types.go -package=mocks -destination=./ratelimit/mocks/types.mock_gen.go
type Limiter interface {
	Limit(ctx context.Context, key string) (bool, error)
}
