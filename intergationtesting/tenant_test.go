package intergationtesting

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ecadlabs/auth/storage"
	uuid "github.com/satori/go.uuid"
)

func givenTenantExists(srv *httptest.Server, name string) (tenant *storage.TenantModel, err error) {
	code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)
	if err != nil {
		return
	}

	model := createTenantModel{Name: name}
	code, tenant, err = createTenant(srv, &model, token)
	if err != nil {
		return
	}

	if code != http.StatusForbidden {
		return
	}
	return
}

func givenUserInviteToTenant(srv *httptest.Server, email string, tenantId uuid.UUID, tokenCh chan string) (err error) {
	code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)
	if err != nil {
		return
	}

	if code != http.StatusOK {
		return
	}

	code, err = inviteTenant(srv, token, fmt.Sprintf("%s", tenantId), email)

	if code != http.StatusNoContent {
		return
	}

	if err != nil {
		return
	}

	inviteToken := <-tokenCh
	code, err = acceptInvite(srv, inviteToken)

	if code != http.StatusNoContent {
		return
	}

	if err != nil {
		return
	}

	return
}

func TestInviteToTenant(t *testing.T) {
	srv, _, token, tokenCh, results := BeforeTest(t)
	firstTenant := results.GetTenantbyName(genTestEmail(3))

	code, err := inviteTenant(srv, token, fmt.Sprintf("%s", firstTenant.ID), genTestEmail(0))

	if code != http.StatusNoContent {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	inviteToken := <-tokenCh
	code, err = acceptInvite(srv, inviteToken)

	if code != http.StatusNoContent {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}
}

func TestNewUserShouldNotBeAbleToInvite(t *testing.T) {
	srv, _, token, tokenCh, _ := BeforeTest(t)
	firstTenant, err := givenTenantExists(srv, "test")

	if err != nil {
		t.Error(err)
		return
	}

	givenUserInviteToTenant(srv, genTestEmail(0), firstTenant.ID, tokenCh)

	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &firstTenant.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	code, err = inviteTenant(srv, token, fmt.Sprintf("%s", firstTenant.ID), genTestEmail(1))

	if code != http.StatusForbidden {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}
}

func TestAdmincanPatchAnyMembership(t *testing.T) {
	srv, _, token, tokenCh, results := BeforeTest(t)
	firstTenant, err := givenTenantExists(srv, "test")

	if err != nil {
		t.Error(err)
		return
	}

	givenUserInviteToTenant(srv, genTestEmail(0), firstTenant.ID, tokenCh)
	firstUser := results.GetUser(genTestEmail(0))

	if firstTenant == nil {
		t.Error("Tenant do not exists")
		return
	}

	if firstUser == nil {
		t.Error("User do not exists")
		return
	}

	code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(code)
		return
	}

	code, err = patchMembership(srv, token, firstUser.ID, firstTenant.ID)

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(code)
		return
	}
}

func TestOwnerShouldBeAbleToInviteInHisOwnTenant(t *testing.T) {
	srv, _, token, tokenCh, results := BeforeTest(t)
	firstTenant := results.GetTenantbyName(genTestEmail(0))
	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &firstTenant.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	code, err = inviteTenant(srv, token, fmt.Sprintf("%s", firstTenant.ID), genTestEmail(1))

	if code != http.StatusNoContent {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	inviteToken := <-tokenCh
	code, err = acceptInvite(srv, inviteToken)

	if code != http.StatusNoContent {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}
}

func TestOwnerShouldNotBeAbleToInviteInOtherTenant(t *testing.T) {
	srv, _, _, _, results := BeforeTest(t)
	rootTenant := results.GetTenantbyName("admin@admin")
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))

	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}

	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if code != http.StatusOK {
		t.Error(code)
		return
	}
	code, err = inviteTenant(srv, token, fmt.Sprintf("%s", rootTenant.ID), genTestEmail(1))
	if code != http.StatusForbidden {
		t.Error(code)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
}

func TestOwnerCantDelegateRoleInOtherTenant(t *testing.T) {
	srv, _, _, _, results := BeforeTest(t)
	rootTenant := results.GetTenantbyName("admin@admin")
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}
	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)
	if code != http.StatusOK {
		t.Error(code)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	code, err = patchMembership(srv, token, results.Users[0].ID, rootTenant.ID)
	if code != http.StatusForbidden {
		t.Error(code)
		return
	}
	if err != nil {
		t.Error(code)
		return
	}
}

func TestRegularUserCantDeleteMembership(t *testing.T) {
	srv, _, _, tokenCh, results := BeforeTest(t)
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}

	givenUserInviteToTenant(srv, genTestEmail(1), tenantWithOwner.ID, tokenCh)
	code, token, _, err := doLogin(srv, genTestEmail(1), testPassword, &tenantWithOwner.ID)

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	user := results.GetUser(genTestEmail(0))
	code, err = DeleteMembership(srv, token, tenantWithOwner.ID, user.ID)

	if code != http.StatusForbidden {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(code)
		return
	}
}

func TestOwnerCanDeleteMembership(t *testing.T) {
	srv, _, _, tokenCh, results := BeforeTest(t)
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}

	givenUserInviteToTenant(srv, genTestEmail(1), tenantWithOwner.ID, tokenCh)

	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	user := results.GetUser(genTestEmail(1))
	code, err = DeleteMembership(srv, token, tenantWithOwner.ID, user.ID)

	if code != http.StatusNoContent {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(code)
		return
	}
}

func TestOwnerCanSeeAllMembership(t *testing.T) {
	srv, _, _, tokenCh, results := BeforeTest(t)
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}

	givenUserInviteToTenant(srv, genTestEmail(1), tenantWithOwner.ID, tokenCh)
	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)
	if code != http.StatusOK {
		t.Error(code)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	_, list, err := getTenantMembershipsList(srv, token, tenantWithOwner.ID, url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != 2 {
		t.Error("Should return 2 membership", len(list))
	}
}

func TestRegularUserCantSeeAllMembership(t *testing.T) {
	srv, _, _, tokenCh, results := BeforeTest(t)
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}

	givenUserInviteToTenant(srv, genTestEmail(4), tenantWithOwner.ID, tokenCh)

	code, token, _, err := doLogin(srv, genTestEmail(4), testPassword, &tenantWithOwner.ID)

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	code, _, err = getTenantMembershipsList(srv, token, tenantWithOwner.ID, url.Values{})
	if code != http.StatusForbidden {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}
}

func TestUserShouldBeAbleToSeeAllHisMembership(t *testing.T) {
	srv, _, _, tokenCh, results := BeforeTest(t)
	tenantWithOwner := results.GetTenantbyName(genTestEmail(1))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}

	givenUserInviteToTenant(srv, genTestEmail(0), tenantWithOwner.ID, tokenCh)
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}
	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)
	if code != http.StatusOK {
		t.Error(code)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	code, list, err := getUserMembershipsList(srv, token, results.GetUser(genTestEmail(0)).ID, url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != 2 {
		t.Error("Should return 2 membership", len(list), code)
	}
}

func TestRegularUserShouldNotBeAbleToSeeOtherMembership(t *testing.T) {
	srv, _, _, _, results := BeforeTest(t)
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}
	code, token, _, err := doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)
	if code != http.StatusOK {
		t.Error(code)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	code, _, err = getUserMembershipsList(srv, token, results.GetUser(genTestEmail(1)).ID, url.Values{})
	if code != http.StatusForbidden {
		t.Error(code)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
}

func TestInvitedUserShouldNotBeAbleToLogin(t *testing.T) {
	srv, _, token, _, results := BeforeTest(t)
	tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
	if tenantWithOwner == nil {
		t.Error("Tenant do not exists")
		return
	}
	code, err := inviteTenant(srv, token, fmt.Sprintf("%s", tenantWithOwner.ID), genTestEmail(3))

	if code != http.StatusNoContent {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	code, token, _, err = doLogin(srv, genTestEmail(3), testPassword, &tenantWithOwner.ID)

	if code != http.StatusForbidden {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}
}

func TestDeleteUserShouldArchiveOrphanTenant(t *testing.T) {
	srv, _, token, _, results := BeforeTest(t)

	user := results.GetUser(genTestEmail(5))

	if user == nil {
		t.Error("User does not exists")
		return
	}

	code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	code, err = deleteUser(srv, token, user.ID)

	if code != http.StatusNoContent {
		t.Error(code)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	results, err = fetchTenantAndUsers(srv, true)
	if err != nil {
		t.Error(err)
		return
	}

	deletedTenant := results.GetTenantbyName(genTestEmail(5))

	if err != nil {
		t.Error(err)
		return
	}

	if deletedTenant != nil {
		t.Error("Tenant should have been archived", deletedTenant.Name, deletedTenant.Archived)
	}
}
