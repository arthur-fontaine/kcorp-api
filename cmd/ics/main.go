package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/arthur-fontaine/kcorp-api/cmd/ics/web"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/league"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
	"github.com/arthur-fontaine/kcorp-api/internal/repository/leagueoflegends"
	"github.com/arthur-fontaine/kcorp-api/internal/repository/valorant"
	"github.com/arthur-fontaine/kcorp-api/internal/usecase/matchservice"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting ICS server...")

	log.Println("Initializing calendar...")
	matchesByLeague, err := getMatchesByLeague()
	if err != nil {
		panic(err)
	}

	go func() {
		// Refresh calendar every 5 minutes
		for {
			time.Sleep(5 * time.Minute)

			matchesByLeague, err = getMatchesByLeague()
			if err != nil {
				log.Println("Error refreshing calendar:", err)
				continue
			}
		}
	}()

	http.HandleFunc("/calendar.ics", func(w http.ResponseWriter, r *http.Request) {
		go monitorEvent("calendar.download")
		w.Header().Set("Content-Type", "text/calendar")
		w.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
		w.Header().Set("Cache-Control", "max-age=300")

		leagues := r.URL.Query()["leagues"]
		log.Println("Loading calendar for leagues:", leagues)
		calSerialized := matchesByLeague.Serialize(leagues...)
		log.Println("Sending calendar with", len(calSerialized), "bytes")

		w.Write([]byte(calSerialized))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		web.Home().Render(context.TODO(), w)
	})

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/ics/web/static"))).ServeHTTP(w, r)
	})

	port := os.Getenv("KCORP_API_ICS_PORT")
	if port == "" {
		port = "9753"
	}
	log.Println("ICS server started on port", port)
	log.Fatal("Error starting ICS server:", http.ListenAndServe(":"+port, nil))
}

type MatchesByLeague struct {
	matches        map[string][]match.Match
	serializeCache map[string]string
}

func getMatchesByLeague() (MatchesByLeague, error) {
	lecRepository, err := leagueoflegends.NewLolMatchRepository(leagueoflegends.LECLeagueID, "en-US")
	if err != nil {
		return MatchesByLeague{}, err
	}

	lflRepository, err := leagueoflegends.NewLolMatchRepository(leagueoflegends.LFLLeagueID, "en-US")
	if err != nil {
		return MatchesByLeague{}, err
	}

	vclRepository, err := valorant.NewValorantMatchRepository(league.League{
		ID:   valorant.VCL2025LeagueID, // TODO: Automatically get the league ID (because it will change really often)
		Name: "VCL",
	})
	if err != nil {
		return MatchesByLeague{}, err
	}

	vctRepository, err := valorant.NewValorantMatchRepository(league.League{
		ID:   valorant.VCTKickoff2025LeagueID, // TODO: Automatically get the league ID (because it will change really often)
		Name: "VCT",
	})
	if err != nil {
		return MatchesByLeague{}, err
	}

	ms := matchservice.NewMatchService([]match.Repository{
		lecRepository,
		lflRepository,
		vclRepository,
		vctRepository,
	})

	matches, err := ms.FindNextMatches()
	if err != nil {
		return MatchesByLeague{}, err
	}

	filteredMatches := make([]match.Match, 0, len(matches))
	for _, m := range matches {
		if strings.Contains(m.HomeTeam.Name, "KC") ||
			strings.Contains(m.AwayTeam.Name, "KC") ||
			strings.Contains(m.HomeTeam.Name, "Karmine") ||
			strings.Contains(m.AwayTeam.Name, "Karmine") {
			filteredMatches = append(filteredMatches, m)
		}
	}

	matchesByLeague := MatchesByLeague{
		matches:        make(map[string][]match.Match),
		serializeCache: make(map[string]string),
	}
	for _, m := range filteredMatches {
		matchesByLeague.matches[m.League.Name] = append(matchesByLeague.matches[m.League.Name], m)
	}

	return matchesByLeague, nil
}

func (m MatchesByLeague) Serialize(leagues ...string) string {
	cacheKey := strings.Join(leagues, ",")

	if calSerialized, ok := m.serializeCache[cacheKey]; ok {
		return calSerialized
	}

	if len(leagues) == 0 {
		leagues = make([]string, 0, len(m.matches))
		for k := range m.matches {
			leagues = append(leagues, k)
		}
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//KCorp//KCorp API//EN")
	cal.SetName("Karmine Corp Calendar")
	cal.SetCalscale("GREGORIAN")
	cal.SetTzid("Europe/Paris")

	for _, league := range leagues {
		matches, ok := m.matches[league]
		if !ok {
			continue
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
	}

	events := cal.Events()
	log.Printf("Calendar initialized with %d events", len(events))

	calSerialized := cal.Serialize()
	m.serializeCache[cacheKey] = calSerialized
	return calSerialized
}

func monitorEvent(eventName string) {
	liteeventsPort := os.Getenv("LITEEVENTS_PORT")
	if liteeventsPort == "" {
		log.Println("LITEEVENTS_PORT is not set, skipping event monitoring")
		return
	}

	liteeventsURL := "http://localhost:" + liteeventsPort + "/api/events"

	type Event struct {
		Namespace string   `json:"namespace"`
		Type      string   `json:"type"`
		Data      struct{} `json:"data"`
	}

	event := Event{
		Namespace: "kcorp-api-ics",
		Type:      eventName,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Println("Error marshaling event data:", err)
		return
	}

	req, err := http.NewRequest("POST", liteeventsURL, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	req.AddCookie(&http.Cookie{
		Name:  "passphrase",
		Value: os.Getenv("LITEEVENTS_PASSPHRASE"),
	})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error response from server:", resp.Status)
	}
}
