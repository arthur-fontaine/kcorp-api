package kameto

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arthur-fontaine/kcorp-api/internal/domain/league"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/team"
	"github.com/arthur-fontaine/kcorp-api/internal/pkg/kametoapi"
)

type kametoMatchRepository struct {
	api         *kametoapi.KametoAPI
	competition string
	league      league.League
}

func NewKametoMatchRepository(competition string, league league.League) (*kametoMatchRepository, error) {
	return &kametoMatchRepository{
		api:         kametoapi.NewKametoAPI(),
		competition: competition,
		league:      league,
	}, nil
}

func (k kametoMatchRepository) FindNextMatches(ctx context.Context) ([]match.Match, error) {
	events, err := k.api.GetGroupA(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]match.Match, 0, len(events))
	for _, event := range events {
		if event.CompetitionName == k.competition {
			startDate, err := time.Parse("2006-01-02T15:04:05.000Z", event.Start)
			if err != nil {
				log.Println(err)
				continue
			}
			var duration time.Duration
			endDate, err := time.Parse("2006-01-02T15:04:05.000Z", event.End)
			if err == nil {
				duration = endDate.Sub(startDate)
			} else {
				log.Println(err)
				duration = time.Hour
			}

			var streamLink string
			if event.StreamLink != "" {
				streamLink = fmt.Sprintf("https://www.twitch.tv/%s", event.StreamLink)
			}

			matches = append(matches, match.Match{
				ID: fmt.Sprintf("kameto-%d", event.Id),
				HomeTeam: team.Team{
					Name: event.TeamHomeName,
				},
				AwayTeam: team.Team{
					Name: event.TeamAwayName,
				},
				League:    k.league,
				StreamURL: streamLink,
				DateTime:  startDate,
				Duration:  duration,
			})
		}
	}

	return matches, nil
}
