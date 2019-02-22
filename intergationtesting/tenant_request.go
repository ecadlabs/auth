package intergationtesting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/ecadlabs/auth/jsonpatch"
	"github.com/ecadlabs/auth/storage"
	uuid "github.com/satori/go.uuid"
)

type createTenantModel struct {
	Name    string      `json:"name"`
	OwnerId interface{} `json:"ownerId"`
}

func createTenant(srv *httptest.Server, tenant *createTenantModel, token string) (int, *storage.TenantModel, error) {
	buf, err := json.Marshal(tenant)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest("POST", srv.URL+"/tenants/", bytes.NewReader(buf))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	var t storage.TenantModel

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&t); err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, &t, nil
}

func getTenant(srv *httptest.Server, token string, uid uuid.UUID) (int, *storage.TenantModel, error) {
	req, err := http.NewRequest("GET", srv.URL+"/tenants/"+uid.String(), nil)
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

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, nil
	}

	var t storage.TenantModel

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&t); err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, &t, nil
}

func deleteMembership(srv *httptest.Server, token string, uid, userID uuid.UUID) (int, error) {
	req, err := http.NewRequest("DELETE", srv.URL+"/tenants/"+uid.String()+"/members/"+userID.String(), nil)
	if err != nil {
		return 0, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func inviteTenant(srv *httptest.Server, token, tenantID string, email string) (int, error) {
	roles := make(storage.Roles, 1)
	roles["regular"] = "true"
	data := struct {
		Email string        `json:"email"`
		Roles storage.Roles `json:"roles"`
	}{
		Email: email,
		Roles: roles,
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(srv.URL+"/tenants/%s/members", tenantID), bytes.NewReader(buf))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)

	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func acceptInvite(srv *httptest.Server, token string) (int, error) {
	data := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(srv.URL+"/tenants/accept_invite"), bytes.NewReader(buf))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)

	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func patchMembership(srv *httptest.Server, token string, userID, tenantID uuid.UUID) (int, error) {
	var p jsonpatch.Patch = []*jsonpatch.Op{
		&jsonpatch.Op{
			Op:    "replace",
			Path:  "/membership_type",
			Value: "owner",
		},
		&jsonpatch.Op{
			Op:    "add",
			Path:  "/roles/owner",
			Value: "true",
		},
		&jsonpatch.Op{
			Op:    "remove",
			Path:  "/roles/regular",
			Value: "false",
		},
	}

	buf, err := json.Marshal(p)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf(srv.URL+"/tenants/%v/members/%v", tenantID, userID), bytes.NewReader(buf))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)

	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func getTenantList(srv *httptest.Server, token string, query url.Values) (int, []*storage.TenantModel, error) {
	tmpURL, err := url.Parse(srv.URL)
	if err != nil {
		return 0, nil, err
	}

	tmpURL.Path = "/tenants/"
	tmpURL.RawQuery = query.Encode()
	reqURL := tmpURL.String()

	result := make([]*storage.TenantModel, 0)

	for {
		req, err := http.NewRequest("GET", reqURL, nil)
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
			Value      []*storage.TenantModel `json:"value"`
			TotalCount int                    `json:"total_count"`
			Next       string                 `json:"next"`
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&res); err != nil {
			return 0, nil, err
		}

		resp.Body.Close()

		if len(res.Value) == 0 {
			break
		}

		reqURL = res.Next
		result = append(result, res.Value...)
	}

	return http.StatusOK, result, nil
}

func getTenantMembershipsList(srv *httptest.Server, token string, tenantID uuid.UUID, query url.Values) (int, []*storage.Membership, error) {
	tmpURL, err := url.Parse(srv.URL)
	if err != nil {
		return 0, nil, err
	}

	tmpURL.Path = fmt.Sprintf("/tenants/%s/members", tenantID)
	tmpURL.RawQuery = query.Encode()
	reqURL := tmpURL.String()

	result := make([]*storage.Membership, 0)

	for {
		req, err := http.NewRequest("GET", reqURL, nil)
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
			Value      []*storage.Membership `json:"value"`
			TotalCount int                   `json:"total_count"`
			Next       string                `json:"next"`
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&res); err != nil {
			return 0, nil, err
		}

		resp.Body.Close()

		if len(res.Value) == 0 {
			break
		}

		reqURL = res.Next
		result = append(result, res.Value...)
	}

	return http.StatusOK, result, nil
}

func getUserMembershipsList(srv *httptest.Server, token string, userID uuid.UUID, query url.Values) (int, []*storage.Membership, error) {
	tmpURL, err := url.Parse(srv.URL)
	if err != nil {
		return 0, nil, err
	}

	tmpURL.Path = fmt.Sprintf("/users/%s/memberships", userID)
	tmpURL.RawQuery = query.Encode()
	reqURL := tmpURL.String()

	result := make([]*storage.Membership, 0)

	for {
		req, err := http.NewRequest("GET", reqURL, nil)
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
			Value      []*storage.Membership `json:"value"`
			TotalCount int                   `json:"total_count"`
			Next       string                `json:"next"`
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&res); err != nil {
			return 0, nil, err
		}

		resp.Body.Close()

		if len(res.Value) == 0 {
			break
		}

		reqURL = res.Next
		result = append(result, res.Value...)
	}

	return http.StatusOK, result, nil
}
