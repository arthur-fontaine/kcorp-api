package strafeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/arthur-fontaine/kcorp-api/internal/pkg/cache"
)

func (s *StrafeAPI) GetCalendar(
	ctx context.Context,
	date time.Time,
	cache cache.Cache,
	game GameId,
) ([]CalendarMatch, error) {
	cacheKey := fmt.Sprintf("calendar-strafe-%d-%d-%d-%d", game, date.Year(), date.Month(), date.Day())
	if cache != nil {
		cachedCalendar, err := cache.Get(cacheKey)
		if err == nil {
			var calendar []CalendarMatch
			if err := json.Unmarshal([]byte(cachedCalendar), &calendar); err == nil {
				log.Printf("Returning cached calendar for %d-%d-%d", date.Year(), date.Month(), date.Day())
				return calendar, nil
			}
		}
	}

	r, err := s.makeRequest(ctx, fmt.Sprintf("v1.7/calendar/%d-%d-%d", date.Year(), date.Month(), date.Day()))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch strafe calendar: %w", err)
	}

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch strafe calendar: %d", r.StatusCode)
	}

	var calendar strafeResponse[[]CalendarMatch]
	if err := json.NewDecoder(r.Body).Decode(&calendar); err != nil {
		return nil, fmt.Errorf("failed to decode strafe calendar: %w", err)
	}

	if game == AllGamesId {
		return calendar.Data, nil
	}

	var filteredCalendar []CalendarMatch
	for _, match := range calendar.Data {
		if match.Game == game {
			filteredCalendar = append(filteredCalendar, match)
		}
	}

	if cache != nil {
		calendarBytes, err := json.Marshal(filteredCalendar)
		if err == nil {
			if err := cache.Set(cacheKey, calendarBytes); err != nil {
				log.Printf("Failed to cache calendar: %v", err)
			}
		}
	}

	return filteredCalendar, nil
}

type CalendarMatch struct {
	Game      GameId `json:"game"`
	Id        int    `json:"id"`
	Home      Team   `json:"home"`
	Away      Team   `json:"away"`
	StartTime string `json:"start_date"`
}

type Team struct {
	Name string `json:"name"`
}
