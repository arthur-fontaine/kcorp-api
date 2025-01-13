package matchservice

import (
	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
)

type matchService struct {
	matchRepositories []match.Repository
}

func NewMatchService(matchRepositories []match.Repository) *matchService {
	return &matchService{
		matchRepositories: matchRepositories,
	}
}
