package matchservice

import (
	"context"
	"log"
	"sync"

	"github.com/arthur-fontaine/kcorp-api/internal/domain/match"
)

func (m matchService) FindNextMatches() ([]match.Match, error) {
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
			matchesChan <- matches
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

	log.Printf("Found %d matches\n", len(allMatches))

	return allMatches, nil
}
