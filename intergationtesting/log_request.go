package intergationtesting

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/ecadlabs/auth/storage"
)

func getLogsList(srv *httptest.Server, token string, query url.Values) (int, []*storage.LogEntry, error) {
	tmpUrl, err := url.Parse(srv.URL)
	if err != nil {
		return 0, nil, err
	}

	tmpUrl.Path = "/logs/"
	tmpUrl.RawQuery = query.Encode()
	reqUrl := tmpUrl.String()

	result := make([]*storage.LogEntry, 0)

	for {
		req, err := http.NewRequest("GET", reqUrl, nil)
		if err != nil {
			return 0, nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := srv.Client().Do(req)
		if err != nil {
			return 0, nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			break
		}

		if resp.StatusCode != http.StatusOK {
			return resp.StatusCode, nil, nil
		}

		var res struct {
			Value      []*storage.LogEntry `json:"value"`
			TotalCount int                 `json:"total_count"`
			Next       string              `json:"next"`
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&res); err != nil {
			return 0, nil, err
		}

		resp.Body.Close()

		if len(res.Value) == 0 {
			break
		}

		reqUrl = res.Next
		result = append(result, res.Value...)
	}

	return http.StatusOK, result, nil
}
