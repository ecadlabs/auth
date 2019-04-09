package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const EMPTY_BODY = ""

var (
	ErrInvalidUrl   = errors.New("Invalid api url")
	ErrUnauthorized = errors.New("Invalid credentials unauthorized")
	ErrUserExists   = errors.New("User already exists")
	ErrIPExists     = errors.New("IP already assigned to a user")
)

var (
	LOGIN             = Endpoint{Method: "GET", URL: "login", OkCode: http.StatusOK}
	CREATE_MEMBERSHIP = Endpoint{Method: "POST", URL: "tenants/%s/members/", OkCode: http.StatusNoContent}
	UPDATE_USER       = Endpoint{Method: "PATCH", URL: "users/%s", OkCode: http.StatusOK}
	CREATE_USER       = Endpoint{Method: "POST", URL: "users/", OkCode: http.StatusCreated}
	CREATE_API_KEY    = Endpoint{Method: "POST", URL: "users/%s/api_keys/", OkCode: http.StatusCreated}
	GET_API_KEY_TOKEN = Endpoint{Method: "GET", URL: "users/%s/api_keys/%s/token", OkCode: http.StatusOK}
)

type Client struct {
	httpClient *http.Client
	apiURL     string
	token      string
}

func New(apiURL string) *Client {
	fmt.Printf(apiURL)
	return &Client{
		httpClient: http.DefaultClient,
		apiURL:     apiURL,
	}
}

func strToReader(str string) *strings.Reader {
	return strings.NewReader(str)
}

func JSONifyWhatever(i interface{}) string {
	jsonb, err := json.Marshal(i)
	if err != nil {
		log.Panic(err)
	}
	return string(jsonb)
}

func structToReader(value interface{}) *strings.Reader {
	return strToReader(JSONifyWhatever(value))
}

func emptyReader() *strings.Reader {
	return strToReader("")
}

func (c *Client) getEndpoint(endpoint Endpoint, params ...interface{}) Endpoint {
	urlWithParams := fmt.Sprintf(endpoint.URL, params...)
	return Endpoint{URL: fmt.Sprintf("%s/%s", c.apiURL, urlWithParams), Method: endpoint.Method, OkCode: endpoint.OkCode}
}

type Endpoint struct {
	Method string
	URL    string
	OkCode int
}

func (c *Client) DoRequest(endpoint Endpoint, value interface{}, errMap map[int]error) (*http.Response, error) {
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, structToReader(value))

	if err != nil {
		return nil, ErrInvalidUrl
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, ErrInvalidUrl
	}

	if res.StatusCode != endpoint.OkCode {
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		err, ok := errMap[res.StatusCode]
		if !ok {
			err = errors.New(res.Status)
		}
		return nil, fmt.Errorf("%s\n%s", bodyString, err.Error())
	}

	return res, nil
}
