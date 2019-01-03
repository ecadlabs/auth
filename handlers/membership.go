package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jsonpatch"
	"github.com/ecadlabs/auth/query"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Memberships struct {
	UserStorage       *storage.Storage
	Storage           *storage.TenantStorage
	MembershipStorage *storage.MembershipStorage
	Timeout           time.Duration
	Enforcer          rbac.Enforcer

	BaseURL     func() string
	TenantsPath string
	UsersPath   string
	AuxLogger   *log.Logger
}

func (t *Memberships) MembershipsUrl(tenantId uuid.UUID) string {
	return fmt.Sprintf("%s%s/members", t.BaseURL()+t.TenantsPath, tenantId)
}

func (m *Memberships) UsersURL() string {
	return m.BaseURL() + m.UsersPath
}

func (t *Memberships) context(r *http.Request) (context.Context, context.CancelFunc) {
	if t.Timeout != 0 {
		return context.WithTimeout(r.Context(), t.Timeout)
	}
	return r.Context(), func() {}
}

func (t *Memberships) PatchMembership(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	ctx, cancel := t.context(r)
	defer cancel()

	tenantId, err := uuid.FromString(mux.Vars(r)["tenantId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	userId, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)

	allowedRoles := []string{permissionTenantsFull, permissionTenantsWrite}

	// Tenant owner are allowed to update member from their own tenant
	if tenantId == member.TenantID {
		allowedRoles = append(allowedRoles, permissionTenantsWriteOwned)
	}

	granted, err := role.IsAnyGranted(allowedRoles...)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	var p jsonpatch.Patch
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ops, err := storage.RoleOpsFromPatch(p)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Check role manipulation permissions
	if len(ops.AddRoles) != 0 || len(ops.RemoveRoles) != 0 {
		tmp := make(map[string]struct{}, len(ops.AddRoles)+len(ops.RemoveRoles))

		for _, r := range ops.AddRoles {
			tmp[permissionDelegatePrefix+r] = struct{}{}
		}

		for _, r := range ops.RemoveRoles {
			tmp[permissionDelegatePrefix+r] = struct{}{}
		}

		perm := make([]string, 0, len(tmp))
		for r := range tmp {
			perm = append(perm, r)
		}

		granted, err := role.IsAllGranted(perm...)
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

	updatedMember, err := t.MembershipStorage.UpdateMembership(ctx, tenantId, userId, ops)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if t.AuxLogger != nil {
		if len(ops.Update) != 0 {
			t.AuxLogger.WithFields(logFields(EvUpdate, member.UserID, userId, r)).WithFields(log.Fields(ops.Update)).Printf("User %v updated account %v in tenant %v", member.UserID, userId, tenantId)
		}

		for _, role := range ops.AddRoles {
			t.AuxLogger.WithFields(logFields(EvAddRole, member.UserID, userId, r)).WithField("role", role).Printf("User %v added role `%s' to account %v in tenant %v", member.UserID, userId, tenantId)
		}

		for _, role := range ops.RemoveRoles {
			t.AuxLogger.WithFields(logFields(EvRemoveRole, member.UserID, userId, r)).WithField("role", role).Printf("User %v removed role `%s' from account %v in tenant %v", member.UserID, role, userId, tenantId)
		}
	}

	utils.JSONResponse(w, http.StatusOK, updatedMember)
}

func (t *Memberships) DeleteMembership(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	ctx, cancel := t.context(r)
	defer cancel()

	tenantId, err := uuid.FromString(mux.Vars(r)["tenantId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	userId, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)

	allowedRoles := []string{permissionTenantsFull, permissionTenantsWrite}

	// Every user are able to delete their own membership
	if userId == member.UserID {
		allowedRoles = append(allowedRoles, permissionWriteSelf)
	}

	// Tenant owner are allowed to delete member from their own tenant
	if tenantId == member.TenantID {
		allowedRoles = append(allowedRoles, permissionTenantsWriteOwned)
	}

	granted, err := role.IsAnyGranted(allowedRoles...)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	err = t.MembershipStorage.DeleteMembership(ctx, tenantId, userId)

	if err != nil {
		utils.JSONErrorResponse(w, err)
	}

	// Log
	if t.AuxLogger != nil {
		t.AuxLogger.WithFields(logFields(EvMembershipDelete, member.UserID, userId, r)).Printf("User %v removed member %v in tenant %v", member.UserID, userId, tenantId)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (t *Memberships) FindTenantMemberships(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	ctx, cancel := t.context(r)
	defer cancel()

	tenantId, err := uuid.FromString(mux.Vars(r)["tenantId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	role, err := t.Enforcer.GetRole(ctx, member.Roles.Get()...)

	allowedRoles := []string{permissionTenantsFull, permissionTenantsRead}

	// Tenant owner are allowed to read member from their own tenant
	if tenantId == member.TenantID {
		allowedRoles = append(allowedRoles, permissionTenantsReadOwned)
	}

	granted, err := role.IsAnyGranted(allowedRoles...)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	// Scope down the request to this particular tenant
	r.Form.Set("tenant_id[eq]", tenantId.String())

	q, err := query.FromValues(r.Form, nil)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	memberships, count, nextQuery, err := t.MembershipStorage.GetMemberships(ctx, q)

	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(memberships) == 0 && !q.TotalCount {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pagination
	res := utils.Paginated{
		Value: memberships,
	}

	if q.TotalCount {
		res.TotalCount = &count
	}

	if nextQuery != nil {
		nextUrl, err := url.Parse(t.MembershipsUrl(tenantId))
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

func (u *Memberships) FindUserMemberships(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	granted, err := role.IsAnyGranted(permissionTenantsFull, permissionRead)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	userId, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if !granted && userId == member.UserID {
		granted, err = role.IsAllGranted(permissionReadSelf)
		if err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), errors.CodeUnknown)
		}
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	// Scope down the request to this particular user
	r.Form.Set("user_id[eq]", userId.String())

	q, err := query.FromValues(r.Form, nil)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	memberships, count, nextQuery, err := u.MembershipStorage.GetMemberships(ctx, q)

	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(memberships) == 0 && !q.TotalCount {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pagination
	res := utils.Paginated{
		Value: memberships,
	}

	if q.TotalCount {
		res.TotalCount = &count
	}

	if nextQuery != nil {
		nextUrl, err := url.Parse(fmt.Sprintf("%s/%s/memberships", u.UsersURL(), userId))
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
