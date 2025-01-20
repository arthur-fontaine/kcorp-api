package valorantapi

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (v *ValorantAPI) GetSchedule(ctx context.Context, leagueId string) (Schedule, error) {
	doc, err := v.getDocFromPage(fmt.Sprintf("/event/matches/%s", leagueId))
	if err != nil {
		return Schedule{}, err
	}

	events := []Event{}
	eventErrors := []error{}

	doc.Find(".match-item").Each(func(i int, s *goquery.Selection) {
		rawTime := strings.Trim(s.Find(".match-item-time").Text(), " \n\t") // " 12:00 PM "
		rawDate := s.Parent().Prev().Text()
		rawDate = strings.Replace(rawDate, "Today", "", -1)
		rawDate = strings.Replace(rawDate, "Tomorrow", "", -1)
		rawDate = strings.Replace(rawDate, "Yesterday", "", -1)
		rawDate = strings.Trim(rawDate, " \n\t") // "Sat, January 11, 2025"
		if strings.Contains(rawDate, "TBD") || strings.Contains(rawTime, "TBD") {
			return
		}
		location, err := time.LoadLocation("CET") // TODO: Move this in a config var
		if err != nil {
			log.Printf("Error loading location: %s", err)
			return
		}
		startTime, err := time.ParseInLocation("Mon, January 2, 2006 3:04 PM MST", fmt.Sprintf("%s %s CET", rawDate, rawTime), location)
		if err != nil {
			eventErrors = append(eventErrors, fmt.Errorf("error parsing time: %s", err))
			return
		}

		teams := []EventTeam{}
		s.Find(".match-item-vs-team").Each(func(i int, s *goquery.Selection) {
			gameWinsStr := strings.Trim(s.Find(".match-item-team-score").Text(), " \n\t")
			var gameWins int
			if gameWinsStr != "" {
				gameWins, err = strconv.Atoi(gameWinsStr)
				if err != nil {
					eventErrors = append(eventErrors, err)
					return
				}
			}

			team := EventTeam{
				Name: strings.Trim(s.Find(".match-item-vs-team-name").Text(), " \n\t"),
				Result: EventResult{
					IsWinner: s.Find(".match-item-vs-team").HasClass("mod-winner"),
					GameWins: gameWins,
				},
			}
			teams = append(teams, team)
		})

		matchUrl, exists := s.Attr("href")
		if !exists {
			eventErrors = append(eventErrors, fmt.Errorf("match url not found"))
			return
		}
		matchId := strings.Split(matchUrl, "/")[1]

		match := EventMatch{
			Id:    matchId,
			Teams: teams,
		}

		event := Event{
			StartTime: startTime,
			Match:     match,
		}

		events = append(events, event)
	})

	err = nil
	if len(eventErrors) > 0 {
		err = fmt.Errorf("errors while parsing events %v", eventErrors)
	}

	return Schedule{
		Events: events,
	}, err
}

type Schedule struct {
	Events []Event
}

type Event struct {
	StartTime time.Time
	Match     EventMatch
}

type EventMatch struct {
	Id    string
	Teams []EventTeam
}

type EventTeam struct {
	Id     string
	Name   string
	Result EventResult
}

type EventResult struct {
	IsWinner bool
	GameWins int
}
