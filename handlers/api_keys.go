package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/middleware"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type newKey struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

func (u *Users) NewAPIKey(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	var req newKey
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if _, err = u.checkWritePermissions(role, storage.AccountService, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	key, err := u.Storage.NewKey(ctx, uid, req.TenantID)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	u.AuxLogger.WithFields(logFields(EvDeleteAPIKey, self.ID, uid, r)).WithFields(log.Fields{"key_id": key.ID, "tenant_id": key.TenantID}).Printf("User %v issued API key for service account %v in tenant in tenant %v", self.ID, uid, key.TenantID)

	utils.JSONResponse(w, http.StatusCreated, key)
}

func (u *Users) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	kid, err := uuid.FromString(mux.Vars(r)["keyId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if _, err = u.checkReadPermissions(role, storage.AccountService, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	key, err := u.Storage.GetKey(ctx, uid, kid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, key)
}

func (u *Users) GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if _, err = u.checkReadPermissions(role, storage.AccountService, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	keys, err := u.Storage.GetKeys(ctx, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(keys) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	utils.JSONResponse(w, http.StatusOK, keys)
}

func (u *Users) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	kid, err := uuid.FromString(mux.Vars(r)["keyId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if _, err = u.checkWritePermissions(role, storage.AccountService, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = u.Storage.DeleteKey(ctx, uid, kid); err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	u.AuxLogger.WithFields(logFields(EvDeleteAPIKey, self.ID, uid, r)).WithField("key_id", kid).Printf("User %v removed API key for service account %v", self.ID, uid)

	w.WriteHeader(http.StatusNoContent)
}

func (u *Users) GetAPIToken(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	kid, err := uuid.FromString(mux.Vars(r)["keyId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	writePerm := true
	if v := r.FormValue("permissions"); v != "" {
		writePerm, _ = strconv.ParseBool(v)
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if _, err = u.checkReadPermissions(role, storage.AccountService, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	key, err := u.Storage.GetKey(ctx, uid, kid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	keyMembership, err := u.Storage.GetMembership(ctx, key.TenantID, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	var keyRole rbac.Role
	if writePerm {
		keyRole, err = u.Enforcer.GetRole(ctx, keyMembership.Roles.Get()...)
		if err != nil {
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	opt := userTokenOptions{
		user: &storage.User{
			ID: uid,
		},
		key:        key,
		membership: keyMembership,
		role:       keyRole,
	}

	if err := u.writeUserToken(w, &opt); err != nil {
		utils.JSONErrorResponse(w, err)
	}
}
