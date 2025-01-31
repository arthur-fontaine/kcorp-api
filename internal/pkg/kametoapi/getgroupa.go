package kametoapi

import (
	"context"
	"encoding/json"
)

func (k *KametoAPI) GetGroupA(ctx context.Context) ([]Event, error) {
	r, err := k.makeRequest(ctx, "/group_a")
	if err != nil {
		return nil, err
	}

	var groupA GroupA
	if err := json.NewDecoder(r.Body).Decode(&groupA); err != nil {
		return nil, err
	}

	return append(groupA.Events, groupA.EventsResult...), nil
}

type GroupA struct {
	Events       []Event `json:"events"`
	EventsResult []Event `json:"events_results"`
}

type Event struct {
	Id              int    `json:"id"`
	Title           string `json:"title"`
	CompetitionName string `json:"competition_name"`
	Start           string `json:"start"`
	End             string `json:"end"`
	StreamLink      string `json:"streamLink"`
	TeamHomeImage   string `json:"team_domicile"`
	TeamAwayImage   string `json:"team_exterieur"`
	TeamHomeName    string `json:"team_name_domicile"`
	TeamAwayName    string `json:"team_name_exterieur"`
}
