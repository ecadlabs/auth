package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jq"
	"github.com/ecadlabs/auth/jsonpatch"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

//Memberships is a handler for memberships
type Memberships struct {
	Storage  storage.MembershipStorage
	Timeout  time.Duration
	Enforcer rbac.Enforcer

	BaseURL     func() string
	TenantsPath string
	UsersPath   string
	AuxLogger   *log.Logger
}

func (m *Memberships) membershipsURL(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s%s/members/", m.BaseURL()+m.TenantsPath, tenantID)
}

func (m *Memberships) usersURL() string {
	return m.BaseURL() + m.UsersPath
}

func (m *Memberships) context(r *http.Request) (context.Context, context.CancelFunc) {
	if m.Timeout != 0 {
		return context.WithTimeout(r.Context(), m.Timeout)
	}
	return r.Context(), func() {}
}

// PatchMembership is an endpoint handler to update membership item
func (m *Memberships) PatchMembership(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	ctx, cancel := m.context(r)
	defer cancel()

	tenantID, err := uuid.FromString(mux.Vars(r)["tenantId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	userID, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	role, err := m.Enforcer.GetRole(ctx, member.Roles.Get()...)

	allowedRoles := []string{permissionTenantsFull, permissionTenantsWrite}

	// Tenant owner are allowed to update member from their own tenant
	if tenantID == member.TenantID {
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

	ops, err := storage.OpsFromPatch(p)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Check role manipulation permissions
	addRoles, removeRoles := ops.Add["roles"], ops.Remove["roles"]

	if len(addRoles) != 0 || len(removeRoles) != 0 {
		tmp := make(map[string]struct{}, len(addRoles)+len(removeRoles))

		for _, r := range addRoles {
			tmp[permissionDelegatePrefix+r] = struct{}{}
		}

		for _, r := range removeRoles {
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

	updatedMember, err := m.Storage.UpdateMembership(ctx, tenantID, userID, ops)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if m.AuxLogger != nil {
		if len(ops.Update) != 0 {
			m.AuxLogger.WithFields(logFields(EvUpdate, member.ID, updatedMember.UserID, r)).WithFields(log.Fields(ops.Update)).Printf("User %v updated account %v in tenant %v", member.UserID, userID, tenantID)
		}

		for _, role := range addRoles {
			m.AuxLogger.WithFields(logFields(EvAddRole, member.ID, updatedMember.ID, r)).WithField("role", role).Printf("User %v added role `%s' to account %v in tenant %v", member.UserID, role, userID, tenantID)
		}

		for _, role := range removeRoles {
			m.AuxLogger.WithFields(logFields(EvRemoveRole, member.ID, updatedMember.ID, r)).WithField("role", role).Printf("User %v removed role `%s' from account %v in tenant %v", member.UserID, role, userID, tenantID)
		}
	}

	utils.JSONResponse(w, http.StatusOK, updatedMember)
}

// DeleteMembership is an endpoint handler to delete membership item
func (m *Memberships) DeleteMembership(w http.ResponseWriter, r *http.Request) {
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	ctx, cancel := m.context(r)
	defer cancel()

	tenantID, err := uuid.FromString(mux.Vars(r)["tenantId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	userID, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	role, err := m.Enforcer.GetRole(ctx, member.Roles.Get()...)

	allowedRoles := []string{permissionTenantsFull, permissionTenantsWrite}

	// Every user are able to delete their own membership
	if userID == member.UserID {
		allowedRoles = append(allowedRoles, permissionWriteSelf)
	}

	// Tenant owner are allowed to delete member from their own tenant
	if tenantID == member.TenantID {
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

	err = m.Storage.DeleteMembership(ctx, tenantID, userID)

	if err != nil {
		utils.JSONErrorResponse(w, err)
	}

	// Log
	if m.AuxLogger != nil {
		m.AuxLogger.WithFields(logFields(EvMembershipDelete, member.ID, userID, r)).Printf("User %v removed member %v in tenant %v", member.UserID, userID, tenantID)
	}

	w.WriteHeader(http.StatusNoContent)
}

// FindTenantMemberships is an endpoint handler to get list of membership item for a particular tenant
func (m *Memberships) FindTenantMemberships(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	ctx, cancel := m.context(r)
	defer cancel()

	tenantID, err := uuid.FromString(mux.Vars(r)["tenantId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	role, err := m.Enforcer.GetRole(ctx, member.Roles.Get()...)

	allowedRoles := []string{permissionTenantsFull, permissionTenantsRead}

	// Tenant owner are allowed to read member from their own tenant
	if tenantID == member.TenantID {
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

	q, err := jq.FromValues(r.Form)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	// Scope down the request to this particular tenant
	node := &jq.EQExpr{
		Key:   "tenant_id",
		Value: tenantID.String(),
	}

	if q.Expr == nil {
		q.Expr = &jq.Expr{Node: node}
	} else {
		q.Expr = &jq.Expr{Node: &jq.ANDExpr{
			q.Expr,
			&jq.Expr{Node: node},
		}}
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	memberships, count, nextQuery, err := m.Storage.GetMemberships(ctx, q)

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
		nextURL, err := url.Parse(m.membershipsURL(tenantID))
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

// FindUserMemberships is an endpoint handler to get list of membership item for a particular user
func (m *Memberships) FindUserMemberships(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	ctx, cancel := m.context(r)
	defer cancel()

	role, err := m.Enforcer.GetRole(ctx, member.Roles.Get()...)
	granted, err := role.IsAnyGranted(permissionTenantsFull, permissionFull, permissionTenantsRead, permissionRead)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeUnknown)
	}

	userID, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if !granted && userID == member.UserID {
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

	q, err := jq.FromValues(r.Form)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	// Scope down the request to this particular user
	node := &jq.EQExpr{
		Key:   "user_id",
		Value: userID.String(),
	}

	if q.Expr == nil {
		q.Expr = &jq.Expr{Node: node}
	} else {
		q.Expr = &jq.Expr{Node: &jq.ANDExpr{
			q.Expr,
			&jq.Expr{Node: node},
		}}
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	memberships, count, nextQuery, err := m.Storage.GetMemberships(ctx, q)

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
		nextURL, err := url.Parse(fmt.Sprintf("%s/%s/memberships/", m.usersURL(), userID))
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
