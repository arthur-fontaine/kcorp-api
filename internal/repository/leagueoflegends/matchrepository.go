package leagueoflegends

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/arthur-fontaine/kcorp-api/internal/domain/league"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/team"
	"github.com/arthur-fontaine/kcorp-api/internal/pkg/leagueoflegendsapi"
)

type leagueOfLegendsMatchRepository struct {
	league league.League
	api    *leagueoflegendsapi.LeagueOfLegendsAPI
}

const (
	LECLeagueID = "98767991302996019"
	LFLLeagueID = "105266103462388553"
)

func NewLolMatchRepository(
	leagueId string,
	lang string,
) (match.Repository, error) {
	api := leagueoflegendsapi.NewLeagueOfLegendsAPI(lang)
	leagues, err := api.GetLeagues(context.Background())
	if err != nil {
		return nil, err
	}

	var l league.League
	for _, l1 := range leagues {
		if l1.ID == leagueId {
			l = league.League{
				ID:   l1.ID,
				Name: normalizeLeagueName(l1.Name),
			}
			break
		}
	}
	if l.ID == "" {
		return nil, league.NewLeagueNotFoundError(leagueId)
	}

	matchRepository := leagueOfLegendsMatchRepository{
		league: l,
		api:    api,
	}
	return matchRepository, nil
}

func (m leagueOfLegendsMatchRepository) FindNextMatches(ctx context.Context) ([]match.Match, error) {
	events, err := m.paginateGetSchedule(ctx, m.league.ID)
	if err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("CET") // TODO: Move this in a config var
	if err != nil {
		log.Printf("Error loading location: %s", err)
		return nil, err
	}

	matches := make([]match.Match, 0)
	for _, event := range events {
		t, err := time.ParseInLocation(time.RFC3339, event.StartTime, location)
		if err != nil {
			return nil, err
		}
		if (len(event.Match.Teams) != 2) || (event.Match.Strategy.Count == 0) {
			continue
		}
		match := match.Match{
			ID:       event.Match.Id,
			DateTime: t,
			Duration: time.Duration(1*event.Match.Strategy.Count) * time.Hour,
			HomeTeam: team.Team{
				ID:   event.Match.Teams[0].Code,
				Name: event.Match.Teams[0].Name,
			},
			AwayTeam: team.Team{
				ID:   event.Match.Teams[1].Code,
				Name: event.Match.Teams[1].Name,
			},
			League: m.league,
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (m leagueOfLegendsMatchRepository) paginateGetSchedule(
	ctx context.Context,
	leagueId string,
) ([]leagueoflegendsapi.Event, error) {
	tokens := make(map[string]struct{})
	var tokenChan = make(chan string)
	var wg sync.WaitGroup

	events := make(map[string]leagueoflegendsapi.Event)

	// Initial token
	wg.Add(1)
	go func() {
		tokenChan <- ""
	}()

	go func() {
		wg.Wait()
		close(tokenChan)
	}()

	for token := range tokenChan {
		if _, ok := tokens[token]; ok {
			wg.Done()
			continue
		}

		schedule, err := m.api.GetSchedule(ctx, leagueId, token)
		if err != nil {
			return nil, err
		}

		for _, event := range schedule.Events {
			events[event.Match.Id] = event
		}

		if schedule.Pages.Newer != "" {
			wg.Add(1)
			go func(token string) {
				tokenChan <- token
			}(schedule.Pages.Newer)
		}

		if schedule.Pages.Older != "" {
			wg.Add(1)
			go func(token string) {
				tokenChan <- token
			}(schedule.Pages.Older)
		}

		tokens[token] = struct{}{}
		wg.Done()
	}

	log.Printf("Found %d events for league %s", len(events), leagueId)

	var eventsSlice []leagueoflegendsapi.Event
	for _, event := range events {
		eventsSlice = append(eventsSlice, event)
	}

	return eventsSlice, nil
}

func normalizeLeagueName(name string) string {
	switch name {
	case "La Ligue Française":
		return "LFL"
	default:
		return name
	}
}
