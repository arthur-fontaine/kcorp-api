package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
	"github.com/arthur-fontaine/kcorp-api/internal/repository/leagueoflegends"
	"github.com/arthur-fontaine/kcorp-api/internal/usecase/matchservice"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting ICS server...")

	log.Println("Initializing calendar...")
	cal, err := getCalendar()
	if err != nil {
		panic(err)
	}

	calSerialized := cal.Serialize()

	go func() {
		// Refresh calendar every 5 minutes
		for {
			time.Sleep(5 * time.Minute)

			cal, err := getCalendar()
			if err != nil {
				log.Println("Error refreshing calendar:", err)
				continue
			}

			calSerialized = cal.Serialize()
		}
	}()

	http.HandleFunc("/calendar.ics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		w.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
		w.Header().Set("Cache-Control", "max-age=300")
		w.Write([]byte(calSerialized))
	})

	log.Println("ICS server started on :9753")
	err = http.ListenAndServe(":9753", nil)
	if err != nil {
		log.Println("Error starting ICS server:", err)
	}
}

func getCalendar() (*ics.Calendar, error) {
	cal := ics.NewCalendar()

	lecRepository, err := leagueoflegends.NewLolMatchRepository("KC", leagueoflegends.LECLeagueID, "en-US")
	if err != nil {
		return nil, err
	}

	lflRepository, err := leagueoflegends.NewLolMatchRepository("KCB", leagueoflegends.LFLLeagueID, "en-US")
	if err != nil {
		return nil, err
	}

	ms := matchservice.NewMatchService([]match.Repository{
		lecRepository,
		lflRepository,
	})

	matches, err := ms.FindNextMatches()
	if err != nil {
		return nil, err
	}

	for _, m := range matches {
		summary := fmt.Sprintf("[%s] %s vs %s", m.League.Name, m.HomeTeam.Name, m.AwayTeam.Name)
		summary = strings.Replace(summary, "La Ligue Française", "LFL", -1)

		event := cal.AddEvent(m.ID)
		event.SetStartAt(m.DateTime)
		event.SetDuration(m.Duration)
		event.SetURL(m.StreamURL)
		event.SetSummary(summary)
	}

	events := cal.Events()
	log.Printf("Calendar initialized with %d events", len(events))

	return cal, nil
}
