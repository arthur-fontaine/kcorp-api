package valorant

import (
	"context"
	"time"

	"github.com/arthur-fontaine/kcorp-api/internal/domain/league"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/team"
	"github.com/arthur-fontaine/kcorp-api/internal/pkg/valorantapi"
)

type valorantMatchRepository struct {
	api    *valorantapi.ValorantAPI
	league league.League
}

const (
	VCL2025LeagueID         = "2315"
	VCTKickoff2025LeagueID  = "2276"
	VCTGC2025Stage1LeagueID = "2354"
)

func NewValorantMatchRepository(
	league league.League,
) (match.Repository, error) {
	api := valorantapi.NewValorantAPI()

	matchRepository := valorantMatchRepository{
		api:    api,
		league: league,
	}
	return matchRepository, nil
}

func (m valorantMatchRepository) FindNextMatches(ctx context.Context) ([]match.Match, error) {
	schedule, err := m.api.GetSchedule(ctx, m.league.ID)
	if err != nil {
		return nil, err
	}

	var matches []match.Match
	for _, event := range schedule.Events {
		matches = append(matches, match.Match{
			ID:     event.Match.Id,
			League: m.league,
			HomeTeam: team.Team{
				Name: event.Match.Teams[0].Name,
			},
			AwayTeam: team.Team{
				Name: event.Match.Teams[1].Name,
			},
			DateTime: event.StartTime,
			Duration: time.Duration(1*3) * time.Hour, // 1*3 is because we suppose that every match is a bo3 (one game is 1 hour)
		})
	}

	return matches, nil
}
