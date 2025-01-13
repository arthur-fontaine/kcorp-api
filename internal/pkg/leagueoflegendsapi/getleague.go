package leagueoflegendsapi

import (
	"context"
	"encoding/json"
)

func (l *LeagueOfLegendsAPI) GetLeagues(ctx context.Context) ([]League, error) {
	r, err := l.makeRequest(ctx, "getLeagues", map[string]string{})
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	var resp apiResponse[leagues]
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return resp.Data.Leagues, nil
}

type leagues struct {
	Leagues []League `json:"leagues"`
}

type League struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Image string `json:"image"`
}
