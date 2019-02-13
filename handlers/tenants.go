package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jsonpatch"
	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/query"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (t *Tenants) tenantsURL() string {
	return t.BaseURL() + t.TenantsPath
}

// CreateTenantModel struct used to create new tenant
type CreateTenantModel struct {
	Name  string `json:"name"`
	Owner string `json:"ownerId"`
}

// Tenants handler for crud operation on tenant resources
type Tenants struct {
	UserStorage       *storage.Storage
	Storage           *storage.TenantStorage
	MembershipStorage *storage.MembershipStorage
	Timeout           time.Duration
	Enforcer          rbac.Enforcer

	BaseURL            func() string
	TokenFactory       *TokenFactory
	TenantsPath        string
	InvitePath         string
	Notifier           notification.Notifier
	AuxLogger          *log.Logger
	TenantInviteMaxAge time.Duration
}

func (t *Tenants) context(r *http.Request) (context.Context, context.CancelFunc) {
	if t.Timeout != 0 {
		return context.WithTimeout(r.Context(), t.Timeout)
	}
	return r.Context(), func() {}
}

// FindTenant is a endpoint handler to find a tenant by id
func (t *Tenants) FindTenant(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)
	granted, err := role.IsAllGranted(permissionTenantsFull)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	tenant, err := t.Storage.GetTenant(ctx, uid, member.UserID, !granted)

	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
		return
	}

	utils.JSONResponse(w, 200, &tenant)
}

// FindTenant is a endpoint handler to delete a tenant
func (t *Tenants) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if !t.canUpdateTenant(role, member, uid) {
		utils.JSONError(w, "", errors.CodeForbidden)
		return
	}

	err = t.Storage.DeleteTenant(ctx, uid)

	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
		return
	}

	// Log
	if t.AuxLogger != nil {
		t.AuxLogger.WithFields(logFields(EvArchiveTenant, member.UserID, uid, r)).Printf("User %v archived tenant %v", member.UserID, uid)
	}

	w.WriteHeader(http.StatusNoContent)
}

// FindTenants is a endpoint handler to get a list of tenants
func (t *Tenants) FindTenants(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)
	granted, err := role.IsAnyGranted(permissionTenantsFull, permissionTenantsRead)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	// Default archived value to false
	defaults := make(map[string][]string)
	defaults["archived[eq]"] = []string{"false"}

	q, err := query.FromValues(r.Form, defaults)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	tenants, count, nextQuery, err := t.Storage.GetTenants(ctx, member.UserID, !granted, q)

	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(tenants) == 0 && !q.TotalCount {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pagination
	res := utils.Paginated{
		Value: tenants,
	}

	if q.TotalCount {
		res.TotalCount = &count
	}

	if nextQuery != nil {
		nextURL, err := url.Parse(t.tenantsURL())
		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}

		nextURL.RawQuery = nextQuery.Values().Encode()
		res.Next = nextURL.String()
	}

	utils.JSONResponse(w, http.StatusOK, &res)
}

// CreateTenant is a endpoint handler to create a new tenant
func (t *Tenants) CreateTenant(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	ctx, cancel := t.context(r)
	defer cancel()
	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)
	granted, err := role.IsAnyGranted(permissionTenantsFull, permissionTenantsWrite, permissionTenantsCreate)

	if !granted || err != nil {
		utils.JSONError(w, "", errors.CodeForbidden)
		return
	}

	var createTenant = CreateTenantModel{}
	if err := json.NewDecoder(r.Body).Decode(&createTenant); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if createTenant.Name == "" {
		utils.JSONErrorResponse(w, errors.ErrTenantName)
		return
	}

	// Only super users can add a tenant with an other owner
	granted, err = role.IsAnyGranted(permissionTenantsFull, permissionTenantsWrite)

	if (!granted || err != nil) && createTenant.Owner != "" {
		utils.JSONError(w, "", errors.CodeForbidden)
		return
	}

	if createTenant.Owner == "" {
		createTenant.Owner = member.UserID.String()
	}

	ownerID, err := uuid.FromString(createTenant.Owner)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	newTenants, err := t.Storage.CreateTenant(ctx, createTenant.Name, ownerID)

	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if t.AuxLogger != nil {
		t.AuxLogger.WithFields(logFields(EvCreateTenant, member.UserID, newTenants.ID, r)).Printf("User %v create tenant %v", member.UserID, newTenants.ID)
	}

	utils.JSONResponse(w, 200, newTenants)
}

func (t *Tenants) canUpdateTenant(role rbac.Role, member *storage.Membership, uid uuid.UUID) bool {
	fullAccess, _ := role.IsAnyGranted(permissionTenantsFull)

	if fullAccess {
		return true
	}

	onlyOwn, _ := role.IsAllGranted(permissionTenantsWriteOwned)

	if !onlyOwn {
		return false
	}
	return member.TenantID == uid && member.MembershipType == storage.OwnerMembership && member.MembershipStatus == storage.ActiveState
}

// UpdateTenant is a endpoint handler to update a tenant resource
func (t *Tenants) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if !t.canUpdateTenant(role, member, uid) {
		utils.JSONError(w, "", errors.CodeForbidden)
		return
	}

	var p jsonpatch.Patch
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ops, err := storage.OpsFromPatch(p)

	tenant, err := t.Storage.PatchTenant(ctx, uid, ops)

	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	if t.AuxLogger != nil {
		if len(ops.Update) != 0 {
			t.AuxLogger.WithFields(logFields(EvUpdateTenant, member.UserID, uid, r)).WithFields(log.Fields(ops.Update)).Printf("User %v updated tenant %v", member.UserID, uid)
		}
	}

	utils.JSONResponse(w, 200, &tenant)
}

func (t *Tenants) inviteToken(user *storage.User, tenantID uuid.UUID) (string, error) {
	return t.TokenFactory.Create(
		jwt.MapClaims{
			"tenant_invite": tenantID,
		},
		user,
		t.InvitePath,
		t.TenantInviteMaxAge,
	)
}

// AcceptInvite is a endpoint handler to accept invite to a tenant
func (t *Tenants) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := t.context(r)
	defer cancel()

	invite := struct {
		Token string `json:"token"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&invite); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	requestToken := invite.Token
	if requestToken == "" {
		log.Error(errors.ErrTokenEmpty)
		utils.JSONErrorResponse(w, errors.ErrTokenEmpty)
	}

	// Verify token
	token, err := t.TokenFactory.Verify(requestToken)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(t.InvitePath, true) {
		utils.JSONErrorResponse(w, errors.ErrAudience)
		return
	}

	// Get tenantId
	tenantIDStr, ok := t.TokenFactory.GetClaim(token, "tenant_invite").(string)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}
	tenantID, err := uuid.FromString(tenantIDStr)

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	updates := make(map[string]interface{})
	updates["membership_status"] = storage.ActiveState

	_, err = t.MembershipStorage.UpdateMembership(ctx, tenantID, id, &storage.RoleOps{
		Update: updates,
	})
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// InviteExistingUser is a endpoint handler to invite a user to a tenant
func (t *Tenants) InviteExistingUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey{}).(*storage.User)
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)
	granted, err := role.IsAllGranted(permissionTenantsFull)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if !t.canUpdateTenant(role, member, uid) {
		utils.JSONError(w, "", errors.CodeForbidden)
		return
	}

	var user struct {
		Email          string        `json:"email"`
		MembershipType string        `json:"type"`
		Roles          storage.Roles `json:"roles"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if user.MembershipType == "" {
		user.MembershipType = storage.MemberMembership
	}

	if len(user.Roles) == 0 {
		utils.JSONErrorResponse(w, errors.ErrRolesEmpty)
		return
	}

	if user.MembershipType != storage.OwnerMembership && user.MembershipType != storage.MemberMembership {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	target, err := t.UserStorage.GetUserByEmail(ctx, storage.AccountRegular, user.Email)

	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, errors.ErrUserNotFound)
		return
	}

	// Check write access
	fullGranted, err := role.IsAnyGranted(permissionTenantsFull)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Check delegating access
	if !fullGranted {
		granted, err := member.CanDelegate(role, user.Roles, permissionDelegatePrefix)

		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
		if !granted {
			utils.JSONErrorResponse(w, errors.ErrForbidden)
			return
		}
	}

	// Create invite token
	token, err := t.inviteToken(target, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	tenant, err := t.Storage.GetTenant(ctx, uid, member.UserID, !granted)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	membership, _ := t.MembershipStorage.GetMembership(ctx, uid, target.ID)

	if membership != nil && membership.MembershipStatus != storage.InvitedState {
		utils.JSONErrorResponse(w, errors.ErrMembershipExisits)
		return
	}

	// Only create membership if it does not exists
	if membership == nil {
		err = t.MembershipStorage.AddMembership(ctx, uid, target, storage.InvitedState, user.MembershipType, user.Roles)
		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	if err = t.Notifier.Notify(ctx, notification.NotificationTenantInvite, &notification.NotificationData{
		Tenant:      tenant,
		CurrentUser: self,
		TargetUser:  target,
		Token:       token,
		TokenMaxAge: t.TokenFactory.SessionMaxAge,
	}); err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
