package kameto

import (
	"context"
	"fmt"
	"log"
	"strings"
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

			var homeTeam team.Team
			var awayTeam team.Team

			if event.Player != "null" {
				homeTeam = team.Team{
					Name: parsePlayerName(event.Player),
				}
			} else {
				homeTeam = team.Team{
					Name: event.TeamHomeName,
				}
				awayTeam = team.Team{
					Name: event.TeamAwayName,
				}
			}

			matches = append(matches, match.Match{
				ID:        fmt.Sprintf("kameto-%d", event.Id),
				HomeTeam:  homeTeam,
				AwayTeam:  awayTeam,
				League:    k.league,
				StreamURL: streamLink,
				DateTime:  startDate,
				Duration:  duration,
			})
		}
	}

	return matches, nil
}

func parsePlayerName(name string) string {
	players := strings.Split(name, ";")
	playersSet := make(map[string]struct{})
	for _, player := range players {
		playersSet[player] = struct{}{}
	}
	players = make([]string, 0, len(playersSet))
	for player := range playersSet {
		players = append(players, normalizePlayerName(player))
	}
	return strings.Join(players, " & ")
}

func normalizePlayerName(name string) string {
	switch name {
	case "KC CANBIZZ":
		return "KC Canbizz"
	case "KC DOUBLE61":
		return "KC Double61"
	case "KC WETJUNGLER":
		return "KC Wet Jungler"
	default:
		return name
	}
}
