package rocketleague

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/arthur-fontaine/kcorp-api/internal/domain/league"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/team"
	"github.com/arthur-fontaine/kcorp-api/internal/pkg/strafeapi"
	"github.com/arthur-fontaine/kcorp-api/internal/repository/cache"
)

type rocketLeagueMatchRepository struct {
	api        *strafeapi.StrafeAPI
	lowestDate time.Time
	cache      cache.Cache
}

func NewRocketLeagueMatchRepository(
	lowestDate time.Time,
	c cache.Cache,
) (match.Repository, error) {
	api := strafeapi.NewStrafeAPI()

	matchRepository := rocketLeagueMatchRepository{
		api:        api,
		lowestDate: lowestDate,
		cache:      c,
	}
	return matchRepository, nil
}

func (m rocketLeagueMatchRepository) FindNextMatches(ctx context.Context) ([]match.Match, error) {
	calendarChan := make(chan strafeapi.CalendarMatch)
	sem := make(chan struct{}, 10) // semaphore to limit concurrency to 10

	go func() {
		for date := m.lowestDate; date.Before(time.Now().AddDate(0, 1, 0)); date = date.AddDate(0, 0, 1) {
			sem <- struct{}{} // acquire a slot
			go func(date time.Time) {
				defer func() { <-sem }() // release the slot

				var c cache.Cache
				if date.Before(time.Now()) {
					c = m.cache
				}

				calendar, err := m.api.GetCalendar(ctx, date, c, strafeapi.RocketLeagueId)
				if err != nil {
					return
				}
				for _, event := range calendar {
					calendarChan <- event
				}
			}(date)
		}

		// wait for all goroutines to finish
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
		close(calendarChan)
	}()

	var matches []match.Match
	for event := range calendarChan {
		date, err := time.Parse(time.RFC3339, event.StartTime)
		if err != nil {
			return nil, err
		}

		matches = append(matches, match.Match{
			ID: strconv.Itoa(event.Id),
			League: league.League{
				Name: "RL",
			},
			HomeTeam: team.Team{
				Name: event.Home.Name,
			},
			AwayTeam: team.Team{
				Name: event.Away.Name,
			},
			DateTime: date,
			Duration: 1,
		})
	}

	log.Printf("Found %d Rocket League matches", len(matches))

	return matches, nil
}
