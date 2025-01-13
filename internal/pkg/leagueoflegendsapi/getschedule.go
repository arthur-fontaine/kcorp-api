package leagueoflegendsapi

import (
	"context"
	"encoding/json"
)

func (l *LeagueOfLegendsAPI) GetSchedule(ctx context.Context, leagueId string, pageToken string) (Schedule, error) {
	params := map[string]string{
		"leagueId": leagueId,
	}
	if pageToken != "" {
		params["pageToken"] = pageToken
	}

	r, err := l.makeRequest(ctx, "getSchedule", params)
	if err != nil {
		return Schedule{}, err
	}

	defer r.Body.Close()

	var resp apiResponse[struct {
		Schedule Schedule `json:"schedule"`
	}]
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return Schedule{}, err
	}

	return resp.Data.Schedule, nil
}

type Schedule struct {
	Updated string `json:"updated"`
	Pages   struct {
		Older string `json:"older"`
		Newer string `json:"newer"`
	} `json:"pages"`
	Events []Event `json:"events"`
}

type Event struct {
	StartTime string     `json:"startTime"`
	BlockName string     `json:"blockName"`
	Match     EventMatch `json:"match"`
	State     EventState `json:"state"`
}

type EventMatch struct {
	Id       string      `json:"id"`
	Strategy Strategy    `json:"strategy"`
	Teams    []EventTeam `json:"teams"`
}

type EventTeam struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Result struct {
		Outcome  GameOutcome `json:"outcome"`
		GameWins int         `json:"gameWins"`
	} `json:"result"`
}

type EventState string

const (
	EventStateScheduled  EventState = "scheduled"
	EventStateInProgress EventState = "inProgress"
	EventStateCompleted  EventState = "completed"
	EventStateCanceled   EventState = "canceled"
)
