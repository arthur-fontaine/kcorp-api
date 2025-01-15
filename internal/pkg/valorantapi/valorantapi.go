package valorantapi

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const (
	vlrGgURL = "https://www.vlr.gg"
)

type ValorantAPI struct {
}

func NewValorantAPI() *ValorantAPI {
	return &ValorantAPI{}
}

func (v *ValorantAPI) getDocFromPage(url string) (*goquery.Document, error) {
	res, err := http.Get(fmt.Sprintf("%s%s", vlrGgURL, url))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
