package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
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
		go monitorEvent("calendar.download")
		w.Header().Set("Content-Type", "text/calendar")
		w.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
		w.Header().Set("Cache-Control", "max-age=300")
		w.Write([]byte(calSerialized))
	})

	port := os.Getenv("KCORP_API_ICS_PORT")
	if port == "" {
		port = "9753"
	}
	log.Println("ICS server started on port", port)
	log.Fatal("Error starting ICS server:", http.ListenAndServe(":"+port, nil))
}

func getCalendar() (*ics.Calendar, error) {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//KCorp//KCorp API//EN")
	cal.SetName("Karmine Corp Calendar")
	cal.SetCalscale("GREGORIAN")
	cal.SetTzid("Europe/Paris")

	lecRepository, err := leagueoflegends.NewLolMatchRepository(leagueoflegends.LECLeagueID, "en-US")
	if err != nil {
		return nil, err
	}

	lflRepository, err := leagueoflegends.NewLolMatchRepository(leagueoflegends.LFLLeagueID, "en-US")
	if err != nil {
		return nil, err
	}

	vclRepository, err := valorant.NewValorantMatchRepository(league.League{
		ID:   valorant.VCL2025LeagueID, // TODO: Automatically get the league ID (because it will change really often)
		Name: "VCL",
	})
	if err != nil {
		return nil, err
	}

	vctRepository, err := valorant.NewValorantMatchRepository(league.League{
		ID:   valorant.VCTKickoff2025LeagueID, // TODO: Automatically get the league ID (because it will change really often)
		Name: "VCT",
	})
	if err != nil {
		return nil, err
	}

	ms := matchservice.NewMatchService([]match.Repository{
		lecRepository,
		lflRepository,
		vclRepository,
		vctRepository,
	})

	matches, err := ms.FindNextMatches([]string{"KCORP Blue Stars", "Karmine Corp", "KC", "KCB", "Karmine Corp Blue"})
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
