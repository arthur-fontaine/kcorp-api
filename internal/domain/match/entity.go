package match

import (
	"time"

	"github.com/arthur-fontaine/kcorp-api/internal/domain/league"
	"github.com/arthur-fontaine/kcorp-api/internal/domain/team"
)

type Match struct {
	ID        string        `json:"id"`
	DateTime  time.Time     `json:"dateTime"`
	Duration  time.Duration `json:"duration"`
	HomeTeam  team.Team     `json:"homeTeam"`
	AwayTeam  team.Team     `json:"awayTeam"`
	League    league.League `json:"league"`
	StreamURL string        `json:"streamURL"`
}
