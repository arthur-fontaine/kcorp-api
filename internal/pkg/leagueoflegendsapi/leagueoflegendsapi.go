package leagueoflegendsapi

import (
	"context"
	"log"
	"net/http"
	"net/url"
)

const (
	baseLoLEsportsURL  = "https://esports-api.lolesports.com"
	baseLolEsportsPath = "/persisted/gw/"
	apiKey             = "0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z"
)

type LeagueOfLegendsAPI struct {
	Lang string
}

func NewLeagueOfLegendsAPI(lang string) *LeagueOfLegendsAPI {
	return &LeagueOfLegendsAPI{
		Lang: lang,
	}
}

func (l *LeagueOfLegendsAPI) makeRequest(ctx context.Context, urlStr string, params map[string]string) (*http.Response, error) {
	params["hl"] = l.Lang

	// completeUrl := baseLoLEsportsURL + "/" + urlStr + "?"
	// for k, v := range params {
	// 	completeUrl += k + "=" + v + "&"
	// }
	// completeUrl = completeUrl[:len(completeUrl)-1]

	u, err := url.Parse(baseLoLEsportsURL)
	if err != nil {
		return nil, err
	}
	u.Path = baseLolEsportsPath + urlStr
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", apiKey)

	log.Printf("Making request to %s and headers %v\n", u.String(), req.Header)

	client := &http.Client{}
	return client.Do(req)
}

type apiResponse[T any] struct {
	Data T `json:"data"`
}
