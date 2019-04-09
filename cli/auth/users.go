package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ecadlabs/auth/jsonpatch"
	uuid "github.com/satori/go.uuid"
)

var createUserErrMapping = map[int]error{
	http.StatusConflict:     ErrUserExists,
	http.StatusUnauthorized: ErrUnauthorized,
}

var addIpErrMapping = map[int]error{
	http.StatusConflict:     ErrIPExists,
	http.StatusUnauthorized: ErrUnauthorized,
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"account_type"`
}

type APIKey struct {
	ID           uuid.UUID `json:"id"`
	MembershipID uuid.UUID `json:"membership_id"`
	UserID       uuid.UUID `json:"user_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Added        time.Time `json:"added"`
}

func (c *Client) CreateUser(user *User) (*User, error) {
	res, err := c.DoRequest(c.getEndpoint(CREATE_USER), user, createUserErrMapping)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody := User{}

	err = json.NewDecoder(res.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

type AddIpRequest struct {
	UserID string
	IPs    []string
}

// TODO: Probably make sense to convert this in an update user endpoint and use it to add ip in layer over it
func (c *Client) AddIp(addIpRequest *AddIpRequest) (*User, error) {
	ops := make([]*jsonpatch.Op, len(addIpRequest.IPs))
	for i, ip := range addIpRequest.IPs {
		ops[i] = &jsonpatch.Op{
			Op:   "add",
			Path: fmt.Sprintf("/address_whitelist/%s", ip),
		}
	}

	res, err := c.DoRequest(c.getEndpoint(UPDATE_USER, addIpRequest.UserID), ops, addIpErrMapping)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody := User{}

	err = json.NewDecoder(res.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (c *Client) CreateApiKey(userID string, tenantId string) (*APIKey, error) {
	newKey := struct {
		TenantID string `json:"tenant_id"`
	}{
		TenantID: tenantId,
	}

	res, err := c.DoRequest(c.getEndpoint(CREATE_API_KEY, userID), newKey, addIpErrMapping)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody := APIKey{}

	err = json.NewDecoder(res.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

type ApiKeyToken struct {
	Token string `json:"token"`
	ID    string `json:"id"`
}

func (c *Client) GetApiKeyToken(userID string, apiKey string) (*ApiKeyToken, error) {
	res, err := c.DoRequest(c.getEndpoint(GET_API_KEY_TOKEN, userID, apiKey), EMPTY_BODY, addIpErrMapping)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var respBody ApiKeyToken

	err = json.NewDecoder(res.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

