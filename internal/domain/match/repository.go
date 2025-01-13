package match

import "context"

type Repository interface {
	FindNextMatches(ctx context.Context) ([]Match, error)
}
