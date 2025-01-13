package league

import (
	"errors"
)

func NewLeagueNotFoundError(leagueID string) error {
	return errors.New("league not found: " + leagueID)
}
