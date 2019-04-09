package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginResponse struct {
	Token string `json: "token"`
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) Login(username string, password string) (*LoginResponse, error) {
	endpoint := c.getEndpoint(LOGIN)
	fmt.Println(endpoint.URL)
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, emptyReader())

	if err != nil {
		fmt.Println(err.Error())
		return nil, ErrInvalidUrl
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(username, password))
	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, ErrUnauthorized
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println(res.Status)
		return nil, ErrUnauthorized
	}

	respBody := LoginResponse{}

	err = json.NewDecoder(res.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return &respBody, nil
}
