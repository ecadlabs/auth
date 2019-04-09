package auth

import (
	"net/http"

	"github.com/ecadlabs/auth/storage"
)

var createMembershipErrMapping = map[int]error{
	http.StatusConflict:     ErrUserExists,
	http.StatusUnauthorized: ErrUnauthorized,
}

type Membership struct {
	UserID   string
	TenantID string
	Role     string
}

func (c *Client) CreateMembership(membership *Membership) error {
	var requestBody = struct {
		ID             string        `json:"id"`
		MembershipType string        `json:"type"`
		BypassInvite   bool          `json:"bypass_invite"`
		Roles          storage.Roles `json:"roles"`
	}{
		ID:             membership.UserID,
		Roles:          storage.Roles{membership.Role: true},
		MembershipType: "member",
		BypassInvite:   true,
	}
	res, err := c.DoRequest(c.getEndpoint(CREATE_MEMBERSHIP, membership.TenantID), requestBody, createMembershipErrMapping)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
