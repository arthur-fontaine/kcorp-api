package matchservice

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
)

func (m matchService) FindNextMatches(acceptableTeamNames []string) ([]match.Match, error) {
	var wg sync.WaitGroup
	matchesChan := make(chan []match.Match, len(m.matchRepositories))
	errChan := make(chan error, len(m.matchRepositories))

	ctx := context.Background()
	for _, repo := range m.matchRepositories {
		wg.Add(1)
		go func(repo match.Repository) {
			matches, err := repo.FindNextMatches(ctx)
			if err != nil {
				wg.Done()
				errChan <- err
				return
			}
			filteredMatches := []match.Match{}
			for _, match := range matches {
				if slices.Contains(acceptableTeamNames, match.HomeTeam.Name) || slices.Contains(acceptableTeamNames, match.AwayTeam.Name) {
					filteredMatches = append(filteredMatches, match)
				}

				match.HomeTeam.Name = normalizeTeamName(match.HomeTeam.Name)
				match.AwayTeam.Name = normalizeTeamName(match.AwayTeam.Name)
			}
			matchesChan <- filteredMatches
			wg.Done()
		}(repo)
	}

	wg.Wait()
	close(matchesChan)
	close(errChan)

	var allMatches []match.Match
	for matches := range matchesChan {
		allMatches = append(allMatches, matches...)
	}

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return allMatches, nil
}

func normalizeTeamName(teamName string) string {
	teamName = strings.Replace(teamName, "KCORP Blue", "KCB", -1)
	teamName = strings.Replace(teamName, "Karmine Corp Blue", "KCB", -1)
	teamName = strings.Replace(teamName, "KCORP Blue Stars", "KCBS", -1)
	teamName = strings.Replace(teamName, "Karmine Corp Blue Stars", "KCBS", -1)
	teamName = strings.Replace(teamName, "Karmine Corp", "KC", -1)
	return teamName
}
