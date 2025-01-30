package strafeapi

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	strafeAPIUrl = "https://flask-api.strafe.com"
	strafeAPIKey = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxMDAwLCJpYXQiOjE2MTE2NTM0MzcuMzMzMDU5fQ.n9StQPQdpNIx3E4FKFntFuzKWolstKJRd-T4LwXmfmo"
)

type GameId int

const (
	AllGamesId        GameId = -1
	RocketLeagueId    GameId = 7
	LeagueOfLegendsId GameId = 2
)

type StrafeAPI struct {
}

func NewStrafeAPI() *StrafeAPI {
	return &StrafeAPI{}
}

func (s *StrafeAPI) makeRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", strafeAPIUrl+"/"+url, nil)
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+strafeAPIKey)

	if err != nil {
		return nil, err
	}

	log.Printf("Making request to %s", url)
	client := http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}
	response, err := client.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("Request to %s timed out, retrying", url)
			return s.makeRequest(ctx, url)
		}

		return nil, err
	}

	if response.StatusCode == 429 {
		retryAfterStr := response.Header.Get("Retry-After")
		if retryAfterStr != "" {
			log.Printf("Rate limited on %s, retrying after %s seconds", url, retryAfterStr)
			retryAfter, err := strconv.Atoi(retryAfterStr)
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(retryAfter) * time.Second)
			return s.makeRequest(ctx, url)
		}
	}

	if !(200 >= response.StatusCode && response.StatusCode <= 299) {
		return nil, fmt.Errorf("request to %s returned %d", url, response.StatusCode)
	}

	return response, nil
}

type strafeResponse[T any] struct {
	Data T `json:"data"`
}
