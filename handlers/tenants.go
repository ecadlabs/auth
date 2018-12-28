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

func (t *Tenants) TenantsUrl() string {
	return t.BaseURL() + t.TenantsPath
}

type CreateTenantModel struct {
	Name string `json:name`
}

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

func (t *Tenants) FindTenant(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, self.Roles.Get()...)
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

	tenant, err := t.Storage.GetTenant(ctx, uid, self, !granted)

	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
		return
	}

	utils.JSONResponse(w, 200, &tenant)
}

func (t *Tenants) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, self.Roles.Get()...)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if !t.CanUpdateTenant(role, self, uid) {
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
		t.AuxLogger.WithFields(logFields(EvArchiveTenant, self.ID, uid, r)).Printf("User %v archived tenant %v", self.ID, uid)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (t *Tenants) FindTenants(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	self := r.Context().Value(UserContextKey).(*storage.User)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, self.Roles.Get()...)
	granted, err := role.IsAllGranted(permissionTenantsFull)

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

	tenants, count, nextQuery, err := t.Storage.GetTenants(ctx, self, !granted, q)

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
		nextUrl, err := url.Parse(t.TenantsUrl())
		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}

		nextUrl.RawQuery = nextQuery.Values().Encode()
		res.Next = nextUrl.String()
	}

	utils.JSONResponse(w, http.StatusOK, &res)
}

func (t *Tenants) CreateTenant(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)

	ctx, cancel := t.context(r)
	defer cancel()
	role, err := t.Enforcer.GetRole(ctx, self.Roles.Get()...)
	granted, err := role.IsAllGranted(permissionTenantsFull)

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

	newTenants, err := t.Storage.CreateTenant(ctx, createTenant.Name)

	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if t.AuxLogger != nil {
		t.AuxLogger.WithFields(logFields(EvCreateTenant, self.ID, newTenants.ID, r)).Printf("User %v create tenant %v", self.ID, newTenants.ID)
	}

	utils.JSONResponse(w, 200, newTenants)
}

func (t *Tenants) CanUpdateTenant(role rbac.Role, self *storage.User, uid uuid.UUID) bool {
	fullAccess, _ := role.IsAnyGranted(permissionTenantsFull)

	if fullAccess {
		return true
	}

	onlyOwn, _ := role.IsAllGranted(permissionTenantsWriteSelf, permissionReadSelf)

	if !onlyOwn {
		return false
	}
	return self.IsOwner(uid)
}

func (t *Tenants) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, self.Roles.Get()...)

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

	if !t.CanUpdateTenant(role, self, uid) {
		utils.JSONError(w, "", errors.CodeForbidden)
		return
	}

	var p jsonpatch.Patch
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ops, err := storage.TenantOpsFromPatch(p)

	tenant, err := t.Storage.PatchTenant(ctx, uid, ops)

	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	if t.AuxLogger != nil {
		if len(ops.Update) != 0 {
			t.AuxLogger.WithFields(logFields(EvUpdateTenant, self.ID, uid, r)).WithFields(log.Fields(ops.Update)).Printf("User %v updated tenant %v", self.ID, uid)
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

type inviteExistingUser struct {
	Email string `json:email`
}

type acceptInvite struct {
	Token string `json:token`
}

func (t *Tenants) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := t.context(r)
	defer cancel()

	invite := acceptInvite{}

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
	tenantIdStr, ok := t.TokenFactory.GetClaim(token, "tenant_invite").(string)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}
	tenantId, err := uuid.FromString(tenantIdStr)

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

	err = t.MembershipStorage.UpdateMembership(ctx, tenantId, id, "active")
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (t *Tenants) InviteExistingUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	ctx, cancel := t.context(r)
	defer cancel()

	role, err := t.Enforcer.GetRole(ctx, self.Roles.Get()...)
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

	if !t.CanUpdateTenant(role, self, uid) {
		utils.JSONError(w, "", errors.CodeForbidden)
		return
	}

	var user inviteExistingUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	target, err := t.UserStorage.GetUserByEmail(ctx, user.Email)

	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, errors.ErrUserNotFound)
		return
	}

	// Create invite token
	token, err := t.inviteToken(target, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	tenant, err := t.Storage.GetTenant(ctx, uid, self, !granted)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	membership, _ := t.MembershipStorage.GetMembership(ctx, uid, target.ID)

	if membership != nil && membership.Membership_status != "invited" {
		utils.JSONErrorResponse(w, errors.ErrMembershipExisits)
		return
	}

	if membership == nil {
		err = t.MembershipStorage.AddMembership(ctx, uid, target, "invited")
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
}
