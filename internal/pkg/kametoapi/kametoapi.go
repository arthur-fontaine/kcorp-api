package kametoapi

import (
	"context"
	"fmt"
	"net/http"
)

const (
	kametoAPIBaseURL = "https://api2.kametotv.fr/karmine"
)

type KametoAPI struct {
	baseURL string
}

func NewKametoAPI() *KametoAPI {
	return &KametoAPI{
		baseURL: kametoAPIBaseURL,
	}
}

func (k *KametoAPI) makeRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s%s", k.baseURL, url), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	return client.Do(req)
}
